package database

import (
	"encoding/json"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	eventsPkg "main/pkg/events"
	snapshotPkg "main/pkg/snapshot"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"
	"time"

	"github.com/lib/pq"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

type Database struct {
	logger zerolog.Logger
	config configPkg.DatabaseConfig
	client DatabaseClient
	mutex  sync.Mutex
}

func NewDatabase(
	logger zerolog.Logger,
	config configPkg.DatabaseConfig,
) *Database {
	return &Database{
		logger: logger.With().Str("component", "state_manager").Logger(),
		config: config,
	}
}

func (d *Database) Init() {
	var db DatabaseClient

	switch d.config.Type {
	case constants.DatabaseTypeSqlite:
		db = d.InitSqliteDatabase()
	case constants.DatabaseTypePostgres:
		db = d.InitPostgresDatabase()
	default:
		d.logger.Panic().Str("type", d.config.Type).Msg("Unsupported database type")
	}

	if err := db.Migrate(); err != nil {
		d.logger.Panic().Err(err).Msg("Failed to migrate database")
	}

	d.SetClient(db)
}

func (d *Database) SetClient(client DatabaseClient) {
	d.client = client
}

func (d *Database) MaybeMutexLock() {
	if d.config.Type == constants.DatabaseTypeSqlite {
		d.mutex.Lock()
	}
}

func (d *Database) MaybeMutexUnlock() {
	if d.config.Type == constants.DatabaseTypeSqlite {
		d.mutex.Unlock()
	}
}

func (d *Database) InsertBlock(chain string, block *types.Block) error {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	signaturesBytes := utils.MustJSONMarshall(block.Signatures)
	validatorsBytes := utils.MustJSONMarshall(block.Validators)

	_, err := d.client.Exec(
		"INSERT INTO blocks (chain, height, time, proposer, signatures, validators) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING",
		chain,
		block.Height,
		block.Time.Unix(),
		block.Proposer,
		signaturesBytes,
		validatorsBytes,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error saving block")
		return err
	}

	return nil
}

func (d *Database) GetAllBlocks(chain string) (map[int64]*types.Block, error) {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	blocks := map[int64]*types.Block{}

	// Getting blocks
	blocksRows, err := d.client.Query(
		"SELECT height, time, proposer, signatures, validators FROM blocks WHERE chain = $1",
		chain,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting all blocks")
		return blocks, err
	}
	defer func() {
		_ = blocksRows.Close()
		_ = blocksRows.Err()
	}()

	for blocksRows.Next() {
		var (
			blockHeight   int64
			blockTime     int64
			blockProposer string
			signaturesRaw []byte
			validatorsRaw []byte
			signatures    = map[string]int32{}
			validators    = map[string]bool{}
		)

		err = blocksRows.Scan(&blockHeight, &blockTime, &blockProposer, &signaturesRaw, &validatorsRaw)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching block data")
			return blocks, err
		}

		if unmarshalErr := json.Unmarshal(signaturesRaw, &signatures); unmarshalErr != nil {
			d.logger.Error().Err(unmarshalErr).Msg("Error unmarshalling signatures")
		}

		if unmarshalErr := json.Unmarshal(validatorsRaw, &validators); unmarshalErr != nil {
			d.logger.Error().Err(unmarshalErr).Msg("Error unmarshalling validators")
		}

		block := &types.Block{
			Height:     blockHeight,
			Time:       time.Unix(blockTime, 0),
			Proposer:   blockProposer,
			Signatures: signatures,
			Validators: validators,
		}
		blocks[block.Height] = block
	}

	return blocks, nil
}

func (d *Database) TrimBlocksBefore(chain string, height int64) error {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	_, err := d.client.Exec(
		"DELETE FROM blocks WHERE height <= $1 AND chain = $2",
		height,
		chain,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error trimming blocks")
		return err
	}

	return nil
}

func (d *Database) GetAllNotifiers(chain string) (*types.Notifiers, error) {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	notifiers := make(types.Notifiers, 0)

	rows, err := d.client.Query(
		"SELECT operator_address, reporter, user_id, user_name FROM notifiers WHERE chain = $1",
		chain,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting all blocks")
		return &notifiers, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	for rows.Next() {
		var (
			operatorAddress string
			reporter        constants.ReporterName
			userID          string
			userName        string
		)

		err = rows.Scan(&operatorAddress, &reporter, &userID, &userName)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching notifier data")
			return &notifiers, err
		}

		newNotifier := &types.Notifier{
			OperatorAddress: operatorAddress,
			Reporter:        reporter,
			UserID:          userID,
			UserName:        userName,
		}

		notifiers = append(notifiers, newNotifier)
	}

	return &notifiers, nil
}

func (d *Database) InsertNotifier(
	chain string,
	operatorAddress string,
	reporter constants.ReporterName,
	userID string,
	userName string,
) error {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	_, err := d.client.Exec(
		"INSERT INTO notifiers (chain, operator_address, reporter, user_id, user_name) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING",
		chain,
		operatorAddress,
		reporter,
		userID,
		userName,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not insert notifier")
		return err
	}

	return nil
}

func (d *Database) RemoveNotifier(
	chain string,
	operatorAddress string,
	reporter constants.ReporterName,
	userID string,
) error {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	_, err := d.client.Exec(
		"DELETE FROM notifiers WHERE operator_address = $1 AND reporter = $2 AND user_id = $3 AND chain = $4",
		operatorAddress,
		reporter,
		userID,
		chain,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not delete notifier")
		return err
	}

	return nil
}
func (d *Database) GetValueByKey(chain string, key string) ([]byte, error) {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	var value []byte

	err := d.client.
		QueryRow("SELECT value FROM data WHERE key = $1 AND chain = $2", key, chain).
		Scan(&value)

	if err != nil {
		d.logger.Error().Err(err).
			Str("chain", chain).
			Str("key", key).
			Msg("Could not get value")
		return value, err
	}

	return value, err
}

func (d *Database) SetValueByKey(chain string, key string, data []byte) error {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	_, err := d.client.Exec(
		"INSERT INTO data (chain, key, value) VALUES ($1, $2, $3) ON CONFLICT (chain, key) DO UPDATE SET value = $3",
		chain,
		key,
		data,
	)
	if err != nil {
		d.logger.Error().Err(err).Str("key", key).Msg("Could not insert value")
		return err
	}

	return nil
}

func (d *Database) GetLastSnapshot(chain string) (*snapshotPkg.Info, error) {
	rawData, err := d.GetValueByKey(chain, "snapshot")
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not get snapshot")
		return nil, err
	}

	var snapshot snapshotPkg.Info
	if unmarshallErr := json.Unmarshal(rawData, &snapshot); unmarshallErr != nil {
		d.logger.Error().Err(unmarshallErr).Msg("Could not unmarshal snapshot")
		return nil, unmarshallErr
	}

	return &snapshot, nil
}

func (d *Database) SetSnapshot(chain string, snapshot *snapshotPkg.Info) error {
	rawData := utils.MustJSONMarshall(snapshot)

	if err := d.SetValueByKey(chain, "snapshot", rawData); err != nil {
		d.logger.Error().Err(err).Msg("Could not save snapshot")
		return err
	}

	return nil
}

func (d *Database) InsertEvent(chain string, height int64, entry types.ReportEvent) error {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	payloadBytes := utils.MustJSONMarshall(entry)

	_, err := d.client.Exec(
		"INSERT INTO events (chain, event, height, validator, payload, time) VALUES ($1, $2, $3, $4, $5, NOW())",
		chain,
		entry.Type(),
		height,
		entry.GetValidator().OperatorAddress,
		payloadBytes,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error saving event")
		return err
	}

	return nil
}

func (d *Database) FindLastEventsByType(
	chain string,
	eventTypes []constants.EventName,
) ([]types.HistoricalEvent, error) {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	events := []types.HistoricalEvent{}

	rows, err := d.client.Query(
		"SELECT event, height, validator, payload, time FROM events WHERE event = any($1) AND chain = $2 ORDER BY time DESC LIMIT $3",
		pq.Array(eventTypes),
		chain,
		constants.LastEventsCount,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting events by operator address and type")
		return events, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	for rows.Next() {
		var (
			validator string
			eventType constants.EventName
			height    int64
			payload   []byte
			eventTime time.Time
		)

		err = rows.Scan(&eventType, &height, &validator, &payload, &eventTime)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching historical event data")
			return events, err
		}

		event := eventsPkg.MapEventTypesToEvent(eventType)
		if unmarshallErr := json.Unmarshal(payload, &event); unmarshallErr != nil {
			d.logger.Error().Err(unmarshallErr).Msg("Could not unmarshal event!")
			return nil, unmarshallErr
		}

		newEvent := types.HistoricalEvent{
			Chain:     chain,
			Type:      eventType,
			Height:    height,
			Validator: validator,
			Time:      eventTime,
			Event:     event,
		}

		events = append(events, newEvent)
	}

	return events, nil
}

func (d *Database) FindLastEventsByValidator(
	chain string,
	operatorAddress string,
) ([]types.HistoricalEvent, error) {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	events := []types.HistoricalEvent{}

	rows, err := d.client.Query(
		"SELECT event, height, validator, payload, time FROM events WHERE validator = $1 AND chain = $2 ORDER BY time DESC LIMIT $3",
		operatorAddress,
		chain,
		constants.LastEventsCount,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting events by operator address and type")
		return events, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	for rows.Next() {
		var (
			validator string
			eventType constants.EventName
			height    int64
			payload   []byte
			eventTime time.Time
		)

		err = rows.Scan(&eventType, &height, &validator, &payload, &eventTime)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching historical event data")
			return events, err
		}

		event := eventsPkg.MapEventTypesToEvent(eventType)
		if unmarshallErr := json.Unmarshal(payload, &event); unmarshallErr != nil {
			d.logger.Error().Err(unmarshallErr).Msg("Could not unmarshal event!")
			return nil, unmarshallErr
		}

		newEvent := types.HistoricalEvent{
			Chain:     chain,
			Type:      eventType,
			Height:    height,
			Validator: validator,
			Time:      eventTime,
			Event:     event,
		}

		events = append(events, newEvent)
	}

	return events, nil
}

func (d *Database) FindAllJailsCount(chain string) ([]types.ValidatorWithJailsCount, error) {
	d.MaybeMutexLock()
	defer d.MaybeMutexUnlock()

	jailsCount := []types.ValidatorWithJailsCount{}

	rows, err := d.client.Query(
		"SELECT validator, count(*) from events where chain = $1 AND event = $2 GROUP BY chain, validator ORDER BY count DESC",
		chain,
		constants.EventValidatorJailed,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting events by operator address and type")
		return jailsCount, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	for rows.Next() {
		var (
			validator string
			count     int
		)

		err = rows.Scan(&validator, &count)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching jails count")
			return jailsCount, err
		}

		jailsCount = append(jailsCount, types.ValidatorWithJailsCount{
			Validator:  validator,
			JailsCount: count,
		})
	}

	return jailsCount, nil
}
