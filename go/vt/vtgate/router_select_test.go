// Copyright 2014, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vtgate

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/youtube/vitess/go/sqltypes"
	querypb "github.com/youtube/vitess/go/vt/proto/query"
	"github.com/youtube/vitess/go/vt/tabletserver/querytypes"
	"github.com/youtube/vitess/go/vt/topo"
	_ "github.com/youtube/vitess/go/vt/vtgate/vindexes"
)

func TestUnsharded(t *testing.T) {
	router, _, _, sbclookup := createRouterEnv()

	_, err := routerExec(router, "select id from music_user_map where id = 1", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select id from music_user_map where id = 1",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbclookup.Queries, wantQueries) {
		t.Errorf("sbclookup.Queries: %+v, want %+v\n", sbclookup.Queries, wantQueries)
	}

	_, err = routerExec(router, "update music_user_map set id = 1", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select id from music_user_map where id = 1",
		BindVariables: map[string]interface{}{},
	}, {
		Sql:           "update music_user_map set id = 1",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbclookup.Queries, wantQueries) {
		t.Errorf("sbclookup.Queries: %+v, want %+v\n", sbclookup.Queries, wantQueries)
	}

	sbclookup.Queries = nil
	_, err = routerExec(router, "delete from music_user_map", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "delete from music_user_map",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbclookup.Queries, wantQueries) {
		t.Errorf("sbclookup.Queries: %+v, want %+v\n", sbclookup.Queries, wantQueries)
	}

	sbclookup.Queries = nil
	_, err = routerExec(router, "insert into music_user_map values (1)", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "insert into music_user_map values (1)",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbclookup.Queries, wantQueries) {
		t.Errorf("sbclookup.Queries: %+v, want %+v\n", sbclookup.Queries, wantQueries)
	}
}

func TestStreamUnsharded(t *testing.T) {
	router, _, _, _ := createRouterEnv()

	sql := "select id from music_user_map where id = 1"
	result, err := routerStream(router, sql)
	if err != nil {
		t.Error(err)
	}
	wantResult := singleRowResult
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestUnshardedFail(t *testing.T) {
	router, _, _, _ := createRouterEnv()

	getSandbox(KsTestUnsharded).SrvKeyspaceMustFail = 1
	_, err := routerExec(router, "select id from music_user_map where id = 1", nil)
	want := "paramsUnsharded: keyspace TestUnsharded fetch error: topo error GetSrvKeyspace"
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: %v, want %v", err, want)
	}

	_, err = routerExec(router, "select id from sharded_table where id = 1", nil)
	want = "unsharded keyspace TestBadSharding has multiple shards"
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: %v, want %v", err, want)
	}
}

func TestStreamUnshardedFail(t *testing.T) {
	router, _, _, _ := createRouterEnv()

	getSandbox(KsTestUnsharded).SrvKeyspaceMustFail = 1
	sql := "select id from music_user_map where id = 1"
	_, err := routerStream(router, sql)
	want := "paramsUnsharded: keyspace TestUnsharded fetch error: topo error GetSrvKeyspace"
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: %v, want %v", err, want)
	}

	sql = "update music_user_map set a = 1 where id = 1"
	_, err = routerStream(router, sql)
	want = `query "update music_user_map set a = 1 where id = 1" cannot be used for streaming`
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: \n%v, want \n%v", err, want)
	}
}

func TestSelectBindvars(t *testing.T) {
	router, sbc1, sbc2, _ := createRouterEnv()

	_, err := routerExec(router, "select id from user where id = :id", map[string]interface{}{
		"id": 1,
	})
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select id from user where id = :id",
		BindVariables: map[string]interface{}{"id": 1},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	if sbc2.Queries != nil {
		t.Errorf("sbc2.Queries: %+v, want nil\n", sbc2.Queries)
	}
	sbc1.Queries = nil

	_, err = routerExec(router, "select id from user where name in (:name1, :name2)", map[string]interface{}{
		"name1": "foo1",
		"name2": "foo2",
	})
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql: "select id from user where name in ::__vals",
		BindVariables: map[string]interface{}{
			"name1":  "foo1",
			"name2":  "foo2",
			"__vals": []interface{}{"foo1", "foo2"},
		},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	sbc1.Queries = nil

	_, err = routerExec(router, "select id from user where name in (:name1, :name2)", map[string]interface{}{
		"name1": []byte("foo1"),
		"name2": []byte("foo2"),
	})
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql: "select id from user where name in ::__vals",
		BindVariables: map[string]interface{}{
			"name1":  []byte("foo1"),
			"name2":  []byte("foo2"),
			"__vals": []interface{}{"foo1", "foo2"},
		},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
}

func TestSelectEqual(t *testing.T) {
	router, sbc1, sbc2, sbclookup := createRouterEnv()

	_, err := routerExec(router, "select id from user where id = 1", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select id from user where id = 1",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	if sbc2.Queries != nil {
		t.Errorf("sbc2.Queries: %+v, want nil\n", sbc2.Queries)
	}
	sbc1.Queries = nil

	_, err = routerExec(router, "select id from user where id = 3", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select id from user where id = 3",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc2.Queries, wantQueries) {
		t.Errorf("sbc2.Queries: %+v, want %+v\n", sbc2.Queries, wantQueries)
	}
	if execCount := sbc1.ExecCount.Get(); execCount != 1 {
		t.Errorf("sbc1.ExecCount: %v, want 1\n", execCount)
	}
	if sbc1.Queries != nil {
		t.Errorf("sbc1.Queries: %+v, want nil\n", sbc1.Queries)
	}
	sbc2.Queries = nil

	_, err = routerExec(router, "select id from user where id = '3'", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select id from user where id = '3'",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc2.Queries, wantQueries) {
		t.Errorf("sbc2.Queries: %+v, want %+v\n", sbc2.Queries, wantQueries)
	}
	if execCount := sbc1.ExecCount.Get(); execCount != 1 {
		t.Errorf("sbc1.ExecCount: %v, want 1\n", execCount)
	}
	if sbc1.Queries != nil {
		t.Errorf("sbc1.Queries: %+v, want nil\n", sbc1.Queries)
	}
	sbc2.Queries = nil

	_, err = routerExec(router, "select id from user where name = 'foo'", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select id from user where name = 'foo'",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql: "select user_id from name_user_map where name = :name",
		BindVariables: map[string]interface{}{
			"name": "foo",
		},
	}}
	if !reflect.DeepEqual(sbclookup.Queries, wantQueries) {
		t.Errorf("sbclookup.Queries: %+v, want %+v\n", sbclookup.Queries, wantQueries)
	}
}

func TestSelectEqualNotFound(t *testing.T) {
	router, _, _, sbclookup := createRouterEnv()

	sbclookup.setResults([]*sqltypes.Result{{}})
	result, err := routerExec(router, "select id from music where id = 1", nil)
	if err != nil {
		t.Error(err)
	}
	wantResult := &sqltypes.Result{}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}

	sbclookup.setResults([]*sqltypes.Result{{}})
	result, err = routerExec(router, "select id from user where name = 'foo'", nil)
	if err != nil {
		t.Error(err)
	}
	wantResult = &sqltypes.Result{}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestStreamSelectEqual(t *testing.T) {
	router, _, _, _ := createRouterEnv()

	sql := "select id from user where id = 1"
	result, err := routerStream(router, sql)
	if err != nil {
		t.Error(err)
	}
	wantResult := singleRowResult
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestSelectEqualFail(t *testing.T) {
	router, _, _, sbclookup := createRouterEnv()
	s := getSandbox("TestRouter")

	_, err := routerExec(router, "select id from user where id = (select count(*) from music)", nil)
	want := "unsupported"
	if err == nil || !strings.HasPrefix(err.Error(), want) {
		t.Errorf("routerExec: %v, must start with %v", err, want)
	}

	_, err = routerExec(router, "select id from user where id = :aa", nil)
	want = "paramsSelectEqual: could not find bind var :aa"
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: %v, want %v", err, want)
	}

	s.SrvKeyspaceMustFail = 1
	_, err = routerExec(router, "select id from user where id = 1", nil)
	want = "paramsSelectEqual: keyspace TestRouter fetch error: topo error GetSrvKeyspace"
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: %v, want %v", err, want)
	}

	sbclookup.mustFailServer = 1
	_, err = routerExec(router, "select id from music where id = 1", nil)
	want = "paramsSelectEqual: lookup.Map"
	if err == nil || !strings.HasPrefix(err.Error(), want) {
		t.Errorf("routerExec: %v, want prefix %v", err, want)
	}

	s.ShardSpec = "80-"
	_, err = routerExec(router, "select id from user where id = 1", nil)
	want = "paramsSelectEqual: KeyspaceId 166b40b44aba4bd6 didn't match any shards"
	if err == nil || !strings.HasPrefix(err.Error(), want) {
		t.Errorf("routerExec: %v, want prefix %v", err, want)
	}
	s.ShardSpec = DefaultShardSpec

	sbclookup.mustFailServer = 1
	_, err = routerExec(router, "select id from user where name = 'foo'", nil)
	want = "paramsSelectEqual: lookup.Map"
	if err == nil || !strings.HasPrefix(err.Error(), want) {
		t.Errorf("routerExec: %v, want prefix %v", err, want)
	}

	s.ShardSpec = "80-"
	_, err = routerExec(router, "select id from user where name = 'foo'", nil)
	want = "paramsSelectEqual: KeyspaceId 166b40b44aba4bd6 didn't match any shards"
	if err == nil || !strings.HasPrefix(err.Error(), want) {
		t.Errorf("routerExec: %v, want prefix %v", err, want)
	}
	s.ShardSpec = DefaultShardSpec
}

func TestSelectIN(t *testing.T) {
	router, sbc1, sbc2, sbclookup := createRouterEnv()

	_, err := routerExec(router, "select id from user where id in (1)", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql: "select id from user where id in ::__vals",
		BindVariables: map[string]interface{}{
			"__vals": []interface{}{int64(1)},
		},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	if sbc2.Queries != nil {
		t.Errorf("sbc2.Queries: %+v, want nil\n", sbc2.Queries)
	}

	sbc1.Queries = nil
	_, err = routerExec(router, "select id from user where id in (1, 3)", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql: "select id from user where id in ::__vals",
		BindVariables: map[string]interface{}{
			"__vals": []interface{}{int64(1)},
		},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql: "select id from user where id in ::__vals",
		BindVariables: map[string]interface{}{
			"__vals": []interface{}{int64(3)},
		},
	}}
	if !reflect.DeepEqual(sbc2.Queries, wantQueries) {
		t.Errorf("sbc2.Queries: %+v, want %+v\n", sbc2.Queries, wantQueries)
	}

	sbc1.Queries = nil
	sbc2.Queries = nil
	_, err = routerExec(router, "select id from user where name = 'foo'", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select id from user where name = 'foo'",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql: "select user_id from name_user_map where name = :name",
		BindVariables: map[string]interface{}{
			"name": "foo",
		},
	}}
	if !reflect.DeepEqual(sbclookup.Queries, wantQueries) {
		t.Errorf("sbclookup.Queries: %+v, want %+v\n", sbclookup.Queries, wantQueries)
	}
}

func TestStreamSelectIN(t *testing.T) {
	router, _, _, sbclookup := createRouterEnv()

	sql := "select id from user where id in (1)"
	result, err := routerStream(router, sql)
	if err != nil {
		t.Error(err)
	}
	wantResult := singleRowResult
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}

	sql = "select id from user where id in (1, 3)"
	result, err = routerStream(router, sql)
	if err != nil {
		t.Error(err)
	}
	wantResult = &sqltypes.Result{
		Fields: singleRowResult.Fields,
		Rows: [][]sqltypes.Value{
			singleRowResult.Rows[0],
			singleRowResult.Rows[0],
		},
		RowsAffected: 2,
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}

	sql = "select id from user where name = 'foo'"
	result, err = routerStream(router, sql)
	if err != nil {
		t.Error(err)
	}
	wantResult = singleRowResult
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}

	wantQueries := []querytypes.BoundQuery{{
		Sql: "select user_id from name_user_map where name = :name",
		BindVariables: map[string]interface{}{
			"name": "foo",
		},
	}}
	if !reflect.DeepEqual(sbclookup.Queries, wantQueries) {
		t.Errorf("sbclookup.Queries: %+v, want %+v\n", sbclookup.Queries, wantQueries)
	}
}

func TestSelectINFail(t *testing.T) {
	router, _, _, _ := createRouterEnv()

	_, err := routerExec(router, "select id from user where id in (:aa)", nil)
	want := "paramsSelectIN: could not find bind var :aa"
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: %v, want %v", err, want)
	}

	getSandbox("TestRouter").SrvKeyspaceMustFail = 1
	_, err = routerExec(router, "select id from user where id in (1)", nil)
	want = "paramsSelectEqual: keyspace TestRouter fetch error: topo error GetSrvKeyspace"
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: %v, want %v", err, want)
	}
}

func TestSelectScatter(t *testing.T) {
	// Special setup: Don't use createRouterEnv.
	s := createSandbox("TestRouter")
	s.VSchema = routerVSchema
	getSandbox(KsTestUnsharded).VSchema = unshardedVSchema
	shards := []string{"-20", "20-40", "40-60", "60-80", "80-a0", "a0-c0", "c0-e0", "e0-"}
	var conns []*sandboxConn
	for _, shard := range shards {
		sbc := &sandboxConn{}
		conns = append(conns, sbc)
		s.MapTestConn(shard, sbc)
	}
	serv := new(sandboxTopo)
	scatterConn := NewScatterConn(nil, topo.Server{}, serv, "", "aa", 1*time.Second, 10, 2*time.Millisecond, 1*time.Millisecond, 24*time.Hour, nil, "")
	router := NewRouter(context.Background(), serv, "aa", "", scatterConn)

	_, err := routerExec(router, "select id from user", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select id from user",
		BindVariables: map[string]interface{}{},
	}}
	for _, conn := range conns {
		if !reflect.DeepEqual(conn.Queries, wantQueries) {
			t.Errorf("conn.Queries = %#v, want %#v", conn.Queries, wantQueries)
		}
	}
}

func TestStreamSelectScatter(t *testing.T) {
	// Special setup: Don't use createRouterEnv.
	s := createSandbox("TestRouter")
	s.VSchema = routerVSchema
	getSandbox(KsTestUnsharded).VSchema = unshardedVSchema
	shards := []string{"-20", "20-40", "40-60", "60-80", "80-a0", "a0-c0", "c0-e0", "e0-"}
	var conns []*sandboxConn
	for _, shard := range shards {
		sbc := &sandboxConn{}
		conns = append(conns, sbc)
		s.MapTestConn(shard, sbc)
	}
	serv := new(sandboxTopo)
	scatterConn := NewScatterConn(nil, topo.Server{}, serv, "", "aa", 1*time.Second, 10, 2*time.Millisecond, 1*time.Millisecond, 24*time.Hour, nil, "")
	router := NewRouter(context.Background(), serv, "aa", "", scatterConn)

	sql := "select id from user"
	result, err := routerStream(router, sql)
	if err != nil {
		t.Error(err)
	}
	wantResult := &sqltypes.Result{
		Fields: singleRowResult.Fields,
		Rows: [][]sqltypes.Value{
			singleRowResult.Rows[0],
			singleRowResult.Rows[0],
			singleRowResult.Rows[0],
			singleRowResult.Rows[0],
			singleRowResult.Rows[0],
			singleRowResult.Rows[0],
			singleRowResult.Rows[0],
			singleRowResult.Rows[0],
		},
		RowsAffected: 8,
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestSelectScatterFail(t *testing.T) {
	// Special setup: Don't use createRouterEnv.
	s := createSandbox("TestRouter")
	s.VSchema = routerVSchema
	getSandbox(KsTestUnsharded).VSchema = unshardedVSchema
	s.SrvKeyspaceMustFail = 1
	shards := []string{"-20", "20-40", "40-60", "60-80", "80-a0", "a0-c0", "c0-e0", "e0-"}
	var conns []*sandboxConn
	for _, shard := range shards {
		sbc := &sandboxConn{}
		conns = append(conns, sbc)
		s.MapTestConn(shard, sbc)
	}
	serv := new(sandboxTopo)
	scatterConn := NewScatterConn(nil, topo.Server{}, serv, "", "aa", 1*time.Second, 10, 2*time.Millisecond, 1*time.Millisecond, 24*time.Hour, nil, "")
	router := NewRouter(context.Background(), serv, "aa", "", scatterConn)

	_, err := routerExec(router, "select id from user", nil)
	want := "paramsSelectScatter: keyspace TestRouter fetch error: topo error GetSrvKeyspace"
	if err == nil || err.Error() != want {
		t.Errorf("routerExec: %v, want %v", err, want)
	}
}

// TODO(sougou): stream and non-stream testing are very similar.
// Could reuse code,
func TestSimpleJoin(t *testing.T) {
	router, sbc1, sbc2, _ := createRouterEnv()
	result, err := routerExec(router, "select u1.id, u2.id from user u1 join user u2 where u1.id = 1 and u2.id = 3", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select u1.id from user as u1 where u1.id = 1",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select u2.id from user as u2 where u2.id = 3",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc2.Queries, wantQueries) {
		t.Errorf("sbc2.Queries: %+v, want %+v\n", sbc2.Queries, wantQueries)
	}
	wantResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			singleRowResult.Fields[0],
			singleRowResult.Fields[0],
		},
		Rows: [][]sqltypes.Value{
			{
				singleRowResult.Rows[0][0],
				singleRowResult.Rows[0][0],
			},
		},
		RowsAffected: 1,
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestSimpleJoinStream(t *testing.T) {
	router, sbc1, sbc2, _ := createRouterEnv()
	result, err := routerStream(router, "select u1.id, u2.id from user u1 join user u2 where u1.id = 1 and u2.id = 3")
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select u1.id from user as u1 where u1.id = 1",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select u2.id from user as u2 where u2.id = 3",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc2.Queries, wantQueries) {
		t.Errorf("sbc2.Queries: %+v, want %+v\n", sbc2.Queries, wantQueries)
	}
	wantResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			singleRowResult.Fields[0],
			singleRowResult.Fields[0],
		},
		Rows: [][]sqltypes.Value{
			{
				singleRowResult.Rows[0][0],
				singleRowResult.Rows[0][0],
			},
		},
		RowsAffected: 0,
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestVarJoin(t *testing.T) {
	router, sbc1, sbc2, _ := createRouterEnv()
	result1 := []*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
		RowsAffected: 1,
		InsertID:     0,
		Rows: [][]sqltypes.Value{{
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("1")),
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("3")),
		}},
	}}
	sbc1.setResults(result1)
	_, err := routerExec(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 1", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select u1.id, u1.col from user as u1 where u1.id = 1",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select u2.id from user as u2 where u2.id = :u1_col",
		BindVariables: map[string]interface{}{},
	}}
	// We have to use string representation because bindvars type is too complex.
	got := fmt.Sprintf("%+v", sbc2.Queries)
	want := "[{Sql:select u2.id from user as u2 where u2.id = :u1_col BindVariables:map[u1_col:3]}]"
	if got != want {
		t.Errorf("sbc2.Queries: %s, want %s\n", got, want)
	}
}

func TestVarJoinStream(t *testing.T) {
	router, sbc1, sbc2, _ := createRouterEnv()
	result1 := []*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
		RowsAffected: 1,
		InsertID:     0,
		Rows: [][]sqltypes.Value{{
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("1")),
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("3")),
		}},
	}}
	sbc1.setResults(result1)
	_, err := routerStream(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 1")
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select u1.id, u1.col from user as u1 where u1.id = 1",
		BindVariables: map[string]interface{}{},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantQueries = []querytypes.BoundQuery{{
		Sql:           "select u2.id from user as u2 where u2.id = :u1_col",
		BindVariables: map[string]interface{}{},
	}}
	// We have to use string representation because bindvars type is too complex.
	got := fmt.Sprintf("%+v", sbc2.Queries)
	want := "[{Sql:select u2.id from user as u2 where u2.id = :u1_col BindVariables:map[u1_col:3]}]"
	if got != want {
		t.Errorf("sbc2.Queries: %s, want %s\n", got, want)
	}
}

func TestLeftJoin(t *testing.T) {
	router, sbc1, sbc2, _ := createRouterEnv()
	result1 := []*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
		RowsAffected: 1,
		InsertID:     0,
		Rows: [][]sqltypes.Value{{
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("1")),
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("3")),
		}},
	}}
	emptyResult := []*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}}
	sbc1.setResults(result1)
	sbc2.setResults(emptyResult)
	result, err := routerExec(router, "select u1.id, u2.id from user u1 left join user u2 on u2.id = u1.col where u1.id = 1", nil)
	if err != nil {
		t.Error(err)
	}
	wantResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			singleRowResult.Fields[0],
			singleRowResult.Fields[0],
		},
		Rows: [][]sqltypes.Value{
			{
				singleRowResult.Rows[0][0],
				{},
			},
		},
		RowsAffected: 1,
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestLeftJoinStream(t *testing.T) {
	router, sbc1, sbc2, _ := createRouterEnv()
	result1 := []*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
		RowsAffected: 1,
		InsertID:     0,
		Rows: [][]sqltypes.Value{{
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("1")),
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("3")),
		}},
	}}
	emptyResult := []*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}}
	sbc1.setResults(result1)
	sbc2.setResults(emptyResult)
	result, err := routerStream(router, "select u1.id, u2.id from user u1 left join user u2 on u2.id = u1.col where u1.id = 1")
	if err != nil {
		t.Error(err)
	}
	wantResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			singleRowResult.Fields[0],
			singleRowResult.Fields[0],
		},
		Rows: [][]sqltypes.Value{
			{
				singleRowResult.Rows[0][0],
				{},
			},
		},
		RowsAffected: 0,
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestEmptyJoin(t *testing.T) {
	router, sbc1, _, _ := createRouterEnv()
	// Empty result requires a field query for the second part of join,
	// which is sent to shard 0.
	sbc1.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}, {
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}})
	result, err := routerExec(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 1", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select u1.id, u1.col from user as u1 where u1.id = 1",
		BindVariables: map[string]interface{}{},
	}, {
		Sql: "select u2.id from user as u2 where 1 != 1",
		BindVariables: map[string]interface{}{
			"u1_col": nil,
		},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"id", sqltypes.Int32},
		},
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestEmptyJoinStream(t *testing.T) {
	router, sbc1, _, _ := createRouterEnv()
	// Empty result requires a field query for the second part of join,
	// which is sent to shard 0.
	sbc1.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}, {
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}})
	result, err := routerStream(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 1")
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select u1.id, u1.col from user as u1 where u1.id = 1",
		BindVariables: map[string]interface{}{},
	}, {
		Sql: "select u2.id from user as u2 where 1 != 1",
		BindVariables: map[string]interface{}{
			"u1_col": nil,
		},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"id", sqltypes.Int32},
		},
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestEmptyJoinRecursive(t *testing.T) {
	router, sbc1, _, _ := createRouterEnv()
	// Make sure it also works recursively.
	sbc1.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}, {
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
	}, {
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}})
	result, err := routerExec(router, "select u1.id, u2.id, u3.id from user u1 join (user u2 join user u3 on u3.id = u2.col) where u1.id = 1", nil)
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select u1.id from user as u1 where u1.id = 1",
		BindVariables: map[string]interface{}{},
	}, {
		Sql:           "select u2.id, u2.col from user as u2 where 1 != 1",
		BindVariables: map[string]interface{}{},
	}, {
		Sql: "select u3.id from user as u3 where 1 != 1",
		BindVariables: map[string]interface{}{
			"u2_col": nil,
		},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries:\n%+v, want\n%+v\n", sbc1.Queries, wantQueries)
	}
	wantResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"id", sqltypes.Int32},
			{"id", sqltypes.Int32},
		},
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestEmptyJoinRecursiveStream(t *testing.T) {
	router, sbc1, _, _ := createRouterEnv()
	// Make sure it also works recursively.
	sbc1.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}, {
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
	}, {
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}})
	result, err := routerStream(router, "select u1.id, u2.id, u3.id from user u1 join (user u2 join user u3 on u3.id = u2.col) where u1.id = 1")
	if err != nil {
		t.Error(err)
	}
	wantQueries := []querytypes.BoundQuery{{
		Sql:           "select u1.id from user as u1 where u1.id = 1",
		BindVariables: map[string]interface{}{},
	}, {
		Sql:           "select u2.id, u2.col from user as u2 where 1 != 1",
		BindVariables: map[string]interface{}{},
	}, {
		Sql: "select u3.id from user as u3 where 1 != 1",
		BindVariables: map[string]interface{}{
			"u2_col": nil,
		},
	}}
	if !reflect.DeepEqual(sbc1.Queries, wantQueries) {
		t.Errorf("sbc1.Queries: %+v, want %+v\n", sbc1.Queries, wantQueries)
	}
	wantResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"id", sqltypes.Int32},
			{"id", sqltypes.Int32},
		},
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("result: %+v, want %+v", result, wantResult)
	}
}

func TestJoinErrors(t *testing.T) {
	router, sbc1, sbc2, _ := createRouterEnv()

	// First query fails
	sbc1.mustFailServer = 1
	_, err := routerExec(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 1", nil)
	want := "error: err"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("err: %v, must start with %s", err, want)
	}

	// Field query fails
	sbc2.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}})
	sbc1.mustFailServer = 1
	_, err = routerExec(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 3", nil)
	want = "error: err"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("err: %v, must start with %s", err, want)
	}

	// Second query fails
	sbc1.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
		Rows: [][]sqltypes.Value{{
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("1")),
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("3")),
		}},
	}})
	sbc2.mustFailServer = 1
	_, err = routerExec(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 1", nil)
	want = "error: err"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("err: %v, must start with %s", err, want)
	}

	// Nested join query fails on get fields
	sbc2.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
	}})
	sbc1.mustFailServer = 1
	_, err = routerExec(router, "select u1.id, u2.id from user u1 join (user u2 join user u3 on u3.id = u2.col) where u1.id = 3", nil)
	want = "error: err"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("err: %v, must start with %s", err, want)
	}

	// Field query fails on stream join
	sbc2.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
		},
	}})
	sbc1.mustFailServer = 1
	_, err = routerStream(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 3")
	want = "error: err"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("err: %v, must start with %s", err, want)
	}

	// Second query fails on stream join
	sbc1.setResults([]*sqltypes.Result{{
		Fields: []*querypb.Field{
			{"id", sqltypes.Int32},
			{"col", sqltypes.Int32},
		},
		Rows: [][]sqltypes.Value{{
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("1")),
			sqltypes.MakeTrusted(sqltypes.Int32, []byte("3")),
		}},
	}})
	sbc2.mustFailServer = 1
	_, err = routerStream(router, "select u1.id, u2.id from user u1 join user u2 on u2.id = u1.col where u1.id = 1")
	want = "error: err"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("err: %v, must start with %s", err, want)
	}
}
