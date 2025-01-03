package database

import (
	"errors"
	configPkg "main/pkg/config"
	"main/pkg/events"
	loggerPkg "main/pkg/logger"
	"main/pkg/snapshot"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDatabaseInsertBlockFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{ExecError: errors.New("custom error")})

	err := database.InsertBlock("chain", &types.Block{
		Height:     123,
		Time:       time.Now(),
		Proposer:   "123",
		Signatures: map[string]int32{},
		Validators: map[string]bool{},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseInsertBlockOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{})

	err := database.InsertBlock("chain", &types.Block{
		Height:     123,
		Time:       time.Now(),
		Proposer:   "123",
		Signatures: map[string]int32{},
		Validators: map[string]bool{},
	})
	require.NoError(t, err)
}

func TestDatabaseTrimBlocksFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{ExecError: errors.New("custom error")})

	err := database.TrimBlocksBefore("chain", 123)
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseTrimBlocksOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{})

	err := database.TrimBlocksBefore("chain", 123)
	require.NoError(t, err)
}

func TestDatabaseInsertNotifierFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{ExecError: errors.New("custom error")})

	err := database.InsertNotifier("chain", "validator", "reporter", "id", "name")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseInsertNotifierOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{})

	err := database.InsertNotifier("chain", "validator", "reporter", "id", "name")
	require.NoError(t, err)
}

func TestDatabaseRemoveNotifierFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{ExecError: errors.New("custom error")})

	err := database.RemoveNotifier("chain", "validator", "reporter", "id")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseRemoveNotifierOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{})

	err := database.RemoveNotifier("chain", "validator", "reporter", "id")
	require.NoError(t, err)
}

func TestDatabaseSetValueByKeyFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{ExecError: errors.New("custom error")})

	err := database.SetValueByKey("chain", "key", []byte("value"))
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseSetValueByKeyOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{})

	err := database.SetValueByKey("chain", "key", []byte("value"))
	require.NoError(t, err)
}

func TestDatabaseSetSnapshotFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{ExecError: errors.New("custom error")})

	err := database.SetSnapshot("chain", &snapshot.Info{
		Height:   123,
		Snapshot: snapshot.Snapshot{},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseSetSnapshotOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{})

	err := database.SetSnapshot("chain", &snapshot.Info{
		Height:   123,
		Snapshot: snapshot.Snapshot{},
	})
	require.NoError(t, err)
}

func TestDatabaseInsertEventFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{ExecError: errors.New("custom error")})

	err := database.InsertEvent("chain", 123, &events.ValidatorGroupChanged{
		Validator: &types.Validator{OperatorAddress: "validator"},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseInsertEventOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{})

	err := database.InsertEvent("chain", 123, &events.ValidatorGroupChanged{
		Validator: &types.Validator{OperatorAddress: "validator"},
	})
	require.NoError(t, err)
}
