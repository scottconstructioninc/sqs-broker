package fakes

import (
	"github.com/cf-platform-eng/sqs-broker/sqlengine"
)

type FakeProvider struct {
	GetSQLEngineCalled    bool
	GetSQLEngineEngine    string
	GetSQLEngineSQLEngine sqlengine.SQLEngine
	GetSQLEngineError     error
}

func (f *FakeProvider) GetSQLEngine(engine string) (sqlengine.SQLEngine, error) {
	f.GetSQLEngineCalled = true
	f.GetSQLEngineEngine = engine

	return f.GetSQLEngineSQLEngine, f.GetSQLEngineError
}
