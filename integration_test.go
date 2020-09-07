// Copyright (c) 2020 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package gogm

import (
	uuid2 "github.com/google/uuid"

	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRawQuery(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	req := require.New(t)

	conf := Config{
		Username:      "neo4j",
		Password:      "password",
		Host:          "0.0.0.0",
		IsCluster:     false,
		Port:          7687,
		PoolSize:      15,
		IndexStrategy: IGNORE_INDEX,
	}

	req.Nil(Init(&conf, &a{}, &b{}, &c{}))

	sess, err := NewSession(false)
	req.Nil(err)

	uuid := uuid2.New().String()

	req.Nil(sess.Save(&a{
		BaseNode: BaseNode{
			UUID: uuid,
		},
	}))

	raw, err := sess.QueryRaw("match (n) where n.uuid=$uuid return n", map[string]interface{}{
		"uuid": uuid,
	})
	req.Nil(err)
	req.NotEmpty(raw)
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	req := require.New(t)

	conf := Config{
		Username:      "neo4j",
		Password:      "changeme",
		Host:          "0.0.0.0",
		IsCluster:     false,
		Port:          7687,
		PoolSize:      15,
		IndexStrategy: IGNORE_INDEX,
	}

	req.Nil(Init(&conf, &a{}, &b{}, &c{}))

	log.Println("opening session")

	log.Println("testIndexManagement")
	testIndexManagement(req)

	sess, err := NewSession(false)
	req.Nil(err)

	log.Println("test save")
	testSave(sess, req)

	req.Nil(sess.PurgeDatabase())

	req.Nil(sess.Close())

	// Test Opening and Closing Session using SessionConfig
	sessConf, err := NewSessionWithConfig(SessionConfig{
		AccessMode: AccessModeRead,
	})
	req.Nil(err)
	req.Nil(sessConf.Close())

	req.Nil(driver.Close())

}

// runs with integration test
func testSave(sess *Session, req *require.Assertions) {
	req.Nil(sess.Begin())
	a2 := &a{
		TestField: "test",
	}

	b2 := &b{
		TestField: "test",
		TestTime:  time.Now().UTC(),
	}

	b3 := &b{
		TestField: "dasdfasd",
	}

	c1 := &c{
		Start: a2,
		End:   b2,
		Test:  "testing",
	}

	a2.SingleSpecA = c1
	a2.ManyA = []*b{b3}
	b2.SingleSpec = c1
	b3.ManyB = a2

	req.Nil(sess.SaveDepth(a2, 5))

	req.Nil(sess.Commit())
	req.Nil(sess.Begin())

	req.EqualValues(map[string]*RelationConfig{
		"SingleSpecA": {
			Ids:          []int64{b2.Id},
			RelationType: Single,
		},
		"ManyA": {
			Ids:          []int64{b3.Id},
			RelationType: Multi,
		},
	}, a2.LoadMap)
	req.EqualValues(map[string]*RelationConfig{
		"SingleSpec": {
			Ids:          []int64{a2.Id},
			RelationType: Single,
		},
	}, b2.LoadMap)
	req.EqualValues(map[string]*RelationConfig{
		"ManyB": {
			Ids:          []int64{a2.Id},
			RelationType: Single,
		},
	}, b3.LoadMap)
	a2.SingleSpecA = nil
	b2.SingleSpec = nil

	req.Nil(sess.SaveDepth(a2, 5))
	req.Nil(sess.Commit())
	req.Nil(a2.SingleSpecA)
	req.Nil(b2.SingleSpec)

	//test save something that isn't connected to anything
	singleSave := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		ManyA:             []*b{},
		MultiA:            []*b{},
		Created:           time.Now().UTC(),
	}

	req.Nil(sess.Begin())
	req.Nil(sess.SaveDepth(singleSave, 1))
	req.Nil(sess.Commit())
}
