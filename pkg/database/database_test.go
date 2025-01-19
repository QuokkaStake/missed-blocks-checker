package database

import (
	"encoding/json"
	"errors"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/events"
	loggerPkg "main/pkg/logger"
	snapshotPkg "main/pkg/snapshot"
	"main/pkg/types"
	"main/pkg/utils"
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

func TestDatabaseGetNotifiersFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT operator_address, reporter, user_id, user_name FROM notifiers").
		WillReturnError(errors.New("custom error"))

	_, err := database.GetAllNotifiers("chain")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseGetNotifiersOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	rows := sqlmock.NewRows([]string{"operator_address", "reporter", "user_id", "user_name"}).
		AddRow("operator1", "telegram", "123", "username").
		AddRow("operator2", "telegram", "123", "username")

	client.Mock.
		ExpectQuery("SELECT operator_address, reporter, user_id, user_name FROM notifiers").
		WillReturnRows(rows)

	result, err := database.GetAllNotifiers("chain")
	require.NoError(t, err)
	require.Equal(t, &types.Notifiers{
		{OperatorAddress: "operator1", Reporter: "telegram", UserID: "123", UserName: "username"},
		{OperatorAddress: "operator2", Reporter: "telegram", UserID: "123", UserName: "username"},
	}, result)
}

func TestDatabaseGetAllBlocksFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT height, time, proposer, signatures, validators FROM blocks").
		WillReturnError(errors.New("custom error"))

	_, err := database.GetAllBlocks("chain")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseGetBlocksFailToUnmarshal(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	rows := sqlmock.NewRows([]string{"height", "time", "proposer", "signatures", "validators"}).
		AddRow("123", time.Now().Unix(), "proposer", "invalid", "invalid")

	client.Mock.
		ExpectQuery("SELECT height, time, proposer, signatures, validators FROM blocks").
		WillReturnRows(rows)

	result, err := database.GetAllBlocks("chain")
	require.NoError(t, err)
	require.Len(t, result, 1)

	block, ok := result[123]
	require.True(t, ok)
	require.Empty(t, block.Validators)
	require.Empty(t, block.Signatures)
}

func TestDatabaseGetBlocksOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{Type: constants.DatabaseTypeSqlite})
	database.SetClient(client)

	blockTime := time.Now().Round(time.Second)

	rows := sqlmock.NewRows([]string{"height", "time", "proposer", "signatures", "validators"}).
		AddRow(
			"123", blockTime.Unix(),
			"proposer",
			utils.MustJSONMarshall(map[string]int32{"validator": 2}),
			utils.MustJSONMarshall(map[string]bool{"validator": true}),
		)

	client.Mock.
		ExpectQuery("SELECT height, time, proposer, signatures, validators FROM blocks").
		WillReturnRows(rows)

	result, err := database.GetAllBlocks("chain")
	require.NoError(t, err)
	require.Len(t, result, 1)

	block, ok := result[123]
	require.True(t, ok)
	require.Equal(t, &types.Block{
		Height:     123,
		Time:       blockTime,
		Proposer:   "proposer",
		Signatures: map[string]int32{"validator": 2},
		Validators: map[string]bool{"validator": true},
	}, block)
}

func TestDatabaseInitSqlite(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{
		Type: constants.DatabaseTypeSqlite,
		Path: "/tmp/missed-blocks-checker-database.sqlite",
	})
	database.Init()

	require.NotNil(t, database.client)
}

func TestDatabaseInitPostgres(t *testing.T) {
	t.Parallel()

	// invalid db connection, won't connect and panic
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{
		Type: constants.DatabaseTypePostgres,
		Path: "postgres://postgres@localhost:8765/missed_blocks_checker?sslmode=disable",
	})
	database.Init()
}

func TestDatabaseInitUnsupported(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	logger := loggerPkg.GetNopLogger()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{
		Type: "invalid",
	})
	database.Init()
}

func TestDatabaseGetEventsByTypeFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT event, height, validator, payload, time FROM events").
		WillReturnError(errors.New("custom error"))

	_, err := database.FindLastEventsByType("chain", constants.GetEventNames())
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseGetEventsByTypeFailToUnmarshal(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT event, height, validator, payload, time FROM events").
		WillReturnRows(sqlmock.
			NewRows([]string{"event", "height", "validator", "payload", "time"}).
			AddRow("test", 123, "test", "test", time.Now()),
		)

	_, err := database.FindLastEventsByType("chain", constants.GetEventNames())
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid character 'e' in literal")
}

func TestDatabaseGetEventsByTypeOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT event, height, validator, payload, time FROM events").
		WillReturnRows(sqlmock.
			NewRows([]string{"event", "height", "validator", "payload", "time"}).
			AddRow(
				constants.EventValidatorActive,
				123,
				"validator",
				utils.MustJSONMarshall(events.ValidatorActive{Validator: &types.Validator{}}),
				time.Now(),
			),
		)

	_, err := database.FindLastEventsByType("chain", constants.GetEventNames())
	require.NoError(t, err)
}

func TestDatabaseGetEventsByValidatorFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT event, height, validator, payload, time FROM events").
		WillReturnError(errors.New("custom error"))

	_, err := database.FindLastEventsByValidator("chain", "validator")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

func TestDatabaseGetEventsByValidatorFailToUnmarshal(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT event, height, validator, payload, time FROM events").
		WillReturnRows(sqlmock.
			NewRows([]string{"event", "height", "validator", "payload", "time"}).
			AddRow("test", 123, "test", "test", time.Now()),
		)

	_, err := database.FindLastEventsByValidator("chain", "validator")
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid character 'e' in literal")
}

func TestDatabaseGetEventsByValidatorOk(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewStubDatabaseClient()
	database := NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(client)

	client.Mock.
		ExpectQuery("SELECT event, height, validator, payload, time FROM events").
		WillReturnRows(sqlmock.
			NewRows([]string{"event", "height", "validator", "payload", "time"}).
			AddRow(
				constants.EventValidatorActive,
				123,
				"validator",
				utils.MustJSONMarshall(events.ValidatorActive{Validator: &types.Validator{}}),
				time.Now(),
			),
		)

	_, err := database.FindLastEventsByValidator("chain", "validator")
	require.NoError(t, err)
}
