package database

import (
	"database/sql"
	"encoding/json"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/snapshot"
	"main/pkg/types"
	migrations "main/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

type Database struct {
	logger zerolog.Logger
	config configPkg.DatabaseConfig
	client *sql.DB
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
	var db *sql.DB

	switch d.config.Type {
	case constants.DatabaseTypeSqlite:
		db = d.InitSqliteDatabase()
	case constants.DatabaseTypePostgres:
		db = d.InitPostgresDatabase()
	default:
		d.logger.Fatal().Str("type", d.config.Type).Msg("Unsupported database type")
	}

	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		d.logger.Fatal().Err(err).Msg("Could not get migrations folder path")
	}

	for _, entry := range entries {
		d.logger.Info().
			Str("name", entry.Name()).
			Msg("Applying sqlite migration")

		content, err := migrations.FS.ReadFile(entry.Name())
		if err != nil {
			d.logger.Fatal().
				Str("name", entry.Name()).
				Err(err).
				Msg("Could not read migration content")
		}

		statement, err := db.Prepare(string(content))
		if err != nil {
			d.logger.Fatal().
				Str("name", entry.Name()).
				Err(err).
				Msg("Could not prepare migration")
		}
		if _, err := statement.Exec(); err != nil {
			d.logger.Fatal().
				Str("name", entry.Name()).
				Err(err).
				Msg("Could not execute migration")
		}

		statement.Close() //nolint:all
	}

	d.client = db
}

func (d *Database) InsertBlock(chain string, block *types.Block) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	signaturesBytes, err := json.Marshal(block.Signatures)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error marshaling signatures")
		return err
	}

	_, err = d.client.Exec(
		"INSERT INTO blocks (chain, height, time, proposer, signatures) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING",
		chain,
		block.Height,
		block.Time.Unix(),
		block.Proposer,
		signaturesBytes,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error saving block")
		return err
	}

	return nil
}

func (d *Database) GetAllBlocks(chain string) (map[int64]*types.Block, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	blocks := map[int64]*types.Block{}

	// Getting blocks
	blocksRows, err := d.client.Query(
		"SELECT height, time, proposer, signatures FROM blocks WHERE chain = $1",
		chain,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting all blocks")
		return blocks, err
	}
	defer func() {
		_ = blocksRows.Close()
		_ = blocksRows.Err() // or modify return value
	}()

	for blocksRows.Next() {
		var (
			blockHeight   int64
			blockTime     int64
			blockProposer string
			signaturesRaw []byte
			signatures    = map[string]int32{}
		)

		err = blocksRows.Scan(&blockHeight, &blockTime, &blockProposer, &signaturesRaw)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching block data")
			return blocks, err
		}

		if err := json.Unmarshal(signaturesRaw, &signatures); err != nil {
			d.logger.Error().Err(err).Msg("Error unmarshalling signatures")
		}

		block := &types.Block{
			Height:     blockHeight,
			Time:       time.Unix(blockTime, 0),
			Proposer:   blockProposer,
			Signatures: signatures,
		}
		blocks[block.Height] = block
	}

	return blocks, nil
}

func (d *Database) TrimBlocksBefore(chain string, height int64) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

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
	d.mutex.Lock()
	defer d.mutex.Unlock()

	notifiers := make(types.Notifiers, 0)

	rows, err := d.client.Query(
		"SELECT operator_address, reporter, notifier FROM notifiers WHERE chain = $1",
		chain,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting all blocks")
		return &notifiers, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err() // or modify return value
	}()

	for rows.Next() {
		var (
			operatorAddress string
			reporter        constants.ReporterName
			notifier        string
		)

		err = rows.Scan(&operatorAddress, &reporter, &notifier)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching notifier data")
			return &notifiers, err
		}

		newNotifier := &types.Notifier{
			OperatorAddress: operatorAddress,
			Reporter:        reporter,
			Notifier:        notifier,
		}

		notifiers = append(notifiers, newNotifier)
	}

	return &notifiers, nil
}

func (d *Database) InsertNotifier(
	chain string,
	operatorAddress string,
	reporter constants.ReporterName,
	notifier string,
) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, err := d.client.Exec(
		"INSERT INTO notifiers (chain, operator_address, reporter, notifier) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
		chain,
		operatorAddress,
		reporter,
		notifier,
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
	notifier string,
) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, err := d.client.Exec(
		"DELETE FROM notifiers WHERE operator_address = $1 AND reporter = $2 AND notifier = $3 AND chain = $4",
		operatorAddress,
		reporter,
		notifier,
		chain,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not delete notifier")
		return err
	}

	return nil
}

func (d *Database) GetAllActiveSets(chain string) (types.HistoricalValidatorsMap, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	activeSets := make(types.HistoricalValidatorsMap, 0)

	rows, err := d.client.Query(
		"SELECT height, validators FROM validators WHERE chain = $1",
		chain,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting all blocks")
		return types.HistoricalValidatorsMap{}, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err() // or modify return value
	}()

	for rows.Next() {
		var (
			height        int64
			validatorsRaw []byte
			validators    types.HistoricalValidators
		)

		err = rows.Scan(&height, &validatorsRaw)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching active set data")
			return nil, err
		}

		if err := json.Unmarshal(validatorsRaw, &validators); err != nil {
			d.logger.Error().Err(err).Msg("Error unmarshalling historical validators")
			return nil, err
		}

		activeSets[height] = validators
	}

	return activeSets, nil
}

func (d *Database) InsertActiveSet(chain string, height int64, activeSet map[string]bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	historicalValidatorsRaw, err := json.Marshal(activeSet)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error marshalling historical validators")
		return err
	}

	_, err = d.client.Exec(
		"INSERT INTO validators (chain, validators, height) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		chain,
		historicalValidatorsRaw,
		height,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error adding historical validators")
		return err
	}

	return nil
}

func (d *Database) TrimActiveSetsBefore(chain string, height int64) error {
	_, err := d.client.Exec(
		"DELETE FROM validators WHERE chain = $1 AND height <= $2",
		chain,
		height,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not trim active set")
		return err
	}

	return nil
}

func (d *Database) GetValueByKey(chain string, key string) ([]byte, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var value []byte

	err := d.client.
		QueryRow("SELECT value FROM data WHERE key = $1 AND chain = $2", key, chain).
		Scan(&value)

	if err != nil {
		d.logger.Error().Err(err).Str("key", key).Msg("Could not get value")
		return value, err
	}

	return value, err
}

func (d *Database) SetValueByKey(chain string, key string, data []byte) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, err := d.client.Exec(
		"INSERT INTO data (chain, key, value) VALUES ($1, $2, $3) ON CONFLICT DO UPDATE SET value = $3",
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

func (d *Database) GetLastSnapshot(chain string) (*snapshot.Info, error) {
	rawData, err := d.GetValueByKey(chain, "snapshot")
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not get snapshot")
		return nil, err
	}

	var snapshot snapshot.Info
	if err := json.Unmarshal(rawData, &snapshot); err != nil {
		d.logger.Error().Err(err).Msg("Could not unmarshal snapshot")
		return nil, err
	}

	return &snapshot, nil
}

func (d *Database) SetSnapshot(chain string, snapshot *snapshot.Info) error {
	rawData, err := json.Marshal(snapshot)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not marshal snapshot")
		return err
	}

	if err := d.SetValueByKey(chain, "snapshot", rawData); err != nil {
		d.logger.Error().Err(err).Msg("Could not save snapshot")
		return err
	}

	return nil
}
