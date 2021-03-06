// Copyright 2019 ScyllaDB
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"context"
	"io"
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/scylladb/gemini"
	"github.com/scylladb/gocqlx/v2/qb"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type cqlStore struct {
	session               *gocql.Session
	schema                *gemini.Schema
	system                string
	maxRetriesMutate      int
	maxRetriesMutateSleep time.Duration

	ops    *prometheus.CounterVec
	logger *zap.Logger
}

func (cs *cqlStore) name() string {
	return cs.system
}

func (cs *cqlStore) mutate(ctx context.Context, builder qb.Builder, ts time.Time, values ...interface{}) (err error) {
	var i int
	for i = 0; i < cs.maxRetriesMutate; i++ {
		err = cs.doMutate(ctx, builder, ts, values...)
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(cs.maxRetriesMutateSleep):
		}
	}
	if err != nil {
		if w := cs.logger.Check(zap.InfoLevel, "failed to apply mutation"); w != nil {
			w.Write(zap.Int("attempts", i), zap.Error(err))
		}
		return
	}

	cs.ops.WithLabelValues(cs.system, opType(builder)).Inc()
	return
}

func (cs *cqlStore) doMutate(ctx context.Context, builder qb.Builder, ts time.Time, values ...interface{}) error {
	query, _ := builder.ToCql()
	tsUsec := ts.UnixNano() / 1000
	/*
		q := cs.session.Query(query, values...).WithContext(ctx).WithTimestamp(tsUsec)
			key, _ := q.GetRoutingKey()
			if len(values) >= 2 {
				v := values[:2]
				s := strings.TrimRight(strings.Repeat("%v,", 2), ",")
				format := fmt.Sprintf("{\nvalues: []interface{}{%s},\nwant: createOne(\"%s\"),\n},\n", s, hex.EncodeToString(key))
				fmt.Printf(format, v...)
			}
			if err := q.Exec(); err != nil {
	*/
	if err := cs.session.Query(query, values...).WithContext(ctx).WithTimestamp(tsUsec).Exec(); err != nil {
		if err == context.DeadlineExceeded {
			if w := cs.logger.Check(zap.DebugLevel, "deadline exceeded for mutation query"); w != nil {
				w.Write(zap.String("system", cs.system), zap.String("query", query), zap.Error(err))
			}
		}
		if !ignore(err) {
			return errors.Wrapf(err, "[cluster = %s, query = '%s']", cs.system, query)
		}
	}
	return nil
}

func (cs *cqlStore) load(ctx context.Context, builder qb.Builder, values []interface{}) (result []map[string]interface{}, err error) {
	query, _ := builder.ToCql()
	iter := cs.session.Query(query, values...).WithContext(ctx).Iter()
	cs.ops.WithLabelValues(cs.system, opType(builder)).Inc()
	defer func() {
		if e := iter.Close(); err != nil {
			if e == context.DeadlineExceeded {
				if w := cs.logger.Check(zap.DebugLevel, "deadline exceeded for load query"); w != nil {
					w.Write(zap.String("system", cs.system), zap.String("query", query), zap.Error(err))
				}
			}
			if !ignore(e) {
				err = multierr.Append(err, errors.Errorf("system failed: %s", e.Error()))
			}
		}
	}()
	result = loadSet(iter)
	return
}

func (cs cqlStore) close() error {
	cs.session.Close()
	return nil
}

func newSession(cluster *gocql.ClusterConfig, out io.Writer) *gocql.Session {
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	if out != nil {
		session.SetTrace(gocql.NewTraceWriter(session, out))
	}
	return session
}

func ignore(err error) bool {
	if err == nil {
		return true
	}
	switch err {
	case context.Canceled, context.DeadlineExceeded:
		return true
	default:
		return false
	}
}

func opType(builder qb.Builder) string {
	switch builder.(type) {
	case *qb.InsertBuilder:
		return "insert"
	case *qb.DeleteBuilder:
		return "delete"
	case *qb.UpdateBuilder:
		return "update"
	case *qb.SelectBuilder:
		return "select"
	case *qb.BatchBuilder:
		return "batch"
	default:
		return "unknown"
	}
}
