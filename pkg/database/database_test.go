package database

import (
	"encoding/json"
	"errors"
	configPkg "main/pkg/config"
	"main/pkg/events"
	loggerPkg "main/pkg/logger"
	snapshotPkg "main/pkg/snapshot"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

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

	err := database.SetSnapshot("chain", &snapshotPkg.Info{
		Height:   123,
		Snapshot: snapshotPkg.Snapshot{},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseSetSnapshotOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&StubDatabaseClient{})

	err := database.SetSnapshot("chain", &snapshotPkg.Info{
		Height:   123,
		Snapshot: snapshotPkg.Snapshot{},
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

func TestDatabaseGetValueByKeyFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT value FROM data").
		WillReturnError(errors.New("custom error"))

	_, err := database.GetValueByKey("chain", "key")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseGetValueByKeyOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT value FROM data").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).FromCSVString("value"))

	result, err := database.GetValueByKey("chain", "key")
	require.NoError(t, err)
	require.Equal(t, []byte("value"), result)
}

func TestDatabaseGetSnapshotFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT value FROM data").
		WillReturnError(errors.New("custom error"))

	_, err := database.GetLastSnapshot("chain")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseGetSnapshotInvalid(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT value FROM data").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).FromCSVString("value"))

	_, err := database.GetLastSnapshot("chain")
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid character 'v' looking for beginning of value")
}

func TestDatabaseGetSnapshotOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	snapshot := &snapshotPkg.Info{
		Height:   123,
		Snapshot: snapshotPkg.Snapshot{},
	}

	snapshotBytes, err := json.Marshal(snapshot)
	require.NoError(t, err)

	client.Mock.
		ExpectQuery("SELECT value FROM data").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(string(snapshotBytes)))

	result, err := database.GetLastSnapshot("chain")
	require.NoError(t, err)
	require.Equal(t, snapshot, result)
}
