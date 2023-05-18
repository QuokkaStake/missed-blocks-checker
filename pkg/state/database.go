package state

import (
	"context"
	"database/sql"
	"encoding/json"
	configPkg "main/pkg/config"
	snapshotPkg "main/pkg/snapshot"
	"main/pkg/types"
	migrations "main/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

type Database struct {
	logger zerolog.Logger
	config *configPkg.Config
	client *sql.DB
	mutex  sync.Mutex
}

func NewDatabase(logger zerolog.Logger, config *configPkg.Config) *Database {
	return &Database{
		logger: logger.With().Str("component", "state_manager").Logger(),
		config: config,
	}
}

func (d *Database) Init() {
	db, err := sql.Open("sqlite3", d.config.DatabaseConfig.Path)

	if err != nil {
		d.logger.Fatal().Err(err).Msg("Could not open sqlite database")
	}

	var version string
	err = db.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		d.logger.Fatal().Err(err).Msg("Could not query sqlite database")
	}

	d.logger.Info().
		Str("version", version).
		Str("path", d.config.DatabaseConfig.Path).
		Msg("sqlite database connected")

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
		defer statement.Close()
		if _, err := statement.Exec(); err != nil {
			d.logger.Fatal().
				Str("name", entry.Name()).
				Err(err).
				Msg("Could not execute migration")
		}
	}

	d.client = db
}

func (d *Database) InsertBlock(block *types.Block) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	ctx := context.Background()
	tx, err := d.client.BeginTx(ctx, nil)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not create a transaction for saving a block")
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO blocks (chain, height, time, proposer) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
		d.config.ChainConfig.Name,
		block.Height,
		block.Time.Unix(),
		block.Proposer,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error saving block")
		if err = tx.Rollback(); err != nil {
			d.logger.Error().Err(err).Msg("Error rolling back transaction")
			return err
		}

		return err
	}

	for key, signature := range block.Signatures {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO signatures (chain, height, validator_address, signature) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
			d.config.ChainConfig.Name,
			block.Height,
			key,
			signature,
		)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error saving signature")
			if err = tx.Rollback(); err != nil {
				d.logger.Error().Err(err).Msg("Error rolling back transaction")
				return err
			}

			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not commit a transaction")
		return err
	}

	return nil
}

func (d *Database) GetAllBlocks() (map[int64]*types.Block, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	blocks := map[int64]*types.Block{}

	// Getting blocks
	blocksRows, err := d.client.Query(
		"SELECT height, time, proposer FROM blocks WHERE chain = $1",
		d.config.ChainConfig.Name,
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
		)

		err = blocksRows.Scan(&blockHeight, &blockTime, &blockProposer)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching block data")
			return blocks, err
		}

		block := &types.Block{
			Height:     blockHeight,
			Time:       time.Unix(blockTime, 0),
			Proposer:   blockProposer,
			Signatures: map[string]int32{},
		}
		blocks[block.Height] = block
	}

	// Fetching signatures
	signaturesRows, err := d.client.Query(
		"SELECT height, validator_address, signature FROM signatures WHERE chain = $1",
		d.config.ChainConfig.Name,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error getting all blocks")
		return blocks, err
	}
	defer func() {
		_ = signaturesRows.Close()
		_ = signaturesRows.Err() // or modify return value
	}()

	for signaturesRows.Next() {
		var (
			signatureHeight int64
			validatorAddr   []byte
			signature       int32
		)

		err = signaturesRows.Scan(&signatureHeight, &validatorAddr, &signature)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching signature data")
			return blocks, err
		}

		_, ok := blocks[signatureHeight]
		if !ok {
			d.logger.Fatal().
				Int64("height", signatureHeight).
				Msg("Got signature for block we do not have, which should never happen.")
		}

		blocks[signatureHeight].Signatures[string(validatorAddr)] = signature
	}

	return blocks, nil
}

func (d *Database) TrimBlocksBefore(height int64) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	ctx := context.Background()
	tx, err := d.client.BeginTx(ctx, nil)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not create a transaction for trimming blocks")
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM blocks WHERE height <= $1 AND chain = $2",
		height,
		d.config.ChainConfig.Name,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error trimming blocks")
		if err = tx.Rollback(); err != nil {
			d.logger.Error().Err(err).Msg("Error rolling back transaction")
			return err
		}

		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM signatures WHERE height <= $1 AND chain = $2",
		height,
		d.config.ChainConfig.Name,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Error trimming signatures")
		if err = tx.Rollback(); err != nil {
			d.logger.Error().Err(err).Msg("Error rolling back transaction")
			return err
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not commit a transaction")
		return err
	}

	return nil
}

func (d *Database) GetAllNotifiers() (*types.Notifiers, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	notifiers := make(types.Notifiers, 0)

	rows, err := d.client.Query(
		"SELECT operator_address, reporter, notifier FROM notifiers WHERE chain = $1",
		d.config.ChainConfig.Name,
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
			reporter        string
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

func (d *Database) InsertNotifier(operatorAddress, reporter, notifier string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, err := d.client.Exec(
		"INSERT INTO notifiers (chain, operator_address, reporter, notifier) VALUES ($1, $2, $3, $2) ON CONFLICT DO NOTHING",
		d.config.ChainConfig.Name,
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

func (d *Database) RemoveNotifier(operatorAddress, reporter, notifier string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, err := d.client.Exec(
		"DELETE FROM notifiers WHERE operator_address = $1 AND reporter = $2 AND notifier = $3 AND chain = $4",
		operatorAddress,
		reporter,
		notifier,
		d.config.ChainConfig.Name,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not delete notifier")
		return err
	}

	return nil
}

func (d *Database) GetAllActiveSets() (types.HistoricalValidatorsMap, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	activeSets := make(types.HistoricalValidatorsMap, 0)

	rows, err := d.client.Query(
		"SELECT height, validator_address FROM validators WHERE chain = $1",
		d.config.ChainConfig.Name,
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
			height           int64
			validatorAddress string
		)

		err = rows.Scan(&height, &validatorAddress)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error fetching active set data")
			return activeSets, err
		}

		if _, ok := activeSets[height]; !ok {
			activeSets[height] = make(map[string]bool, 0)
		}

		activeSets[height][validatorAddress] = true
	}

	return activeSets, nil
}

func (d *Database) InsertActiveSet(height int64, activeSet map[string]bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	ctx := context.Background()
	tx, err := d.client.BeginTx(ctx, nil)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not create a transaction for inserting active set")
		return err
	}

	for validator := range activeSet {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO validators (chain, validator_address, height) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			d.config.ChainConfig.Name,
			validator,
			height,
		)
		if err != nil {
			d.logger.Error().Err(err).Msg("Error adding active set")
			if err = tx.Rollback(); err != nil {
				d.logger.Error().Err(err).Msg("Error rolling back transaction")
				return err
			}

			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not commit a transaction")
		return err
	}

	return nil
}

func (d *Database) TrimActiveSetsBefore(height int64) error {
	_, err := d.client.Exec(
		"DELETE FROM validators WHERE chain = $1 AND height <= $2",
		d.config.ChainConfig.Name,
		height,
	)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not trim active set")
		return err
	}

	return nil
}

func (d *Database) GetValueByKey(key string) ([]byte, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var value []byte

	err := d.client.
		QueryRow("SELECT value FROM data WHERE key = $1 AND chain = $2", key, d.config.ChainConfig.Name).
		Scan(&value)

	if err != nil {
		d.logger.Error().Err(err).Str("key", key).Msg("Could not get value")
		return value, err
	}

	return value, err
}

func (d *Database) SetValueByKey(key string, data []byte) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, err := d.client.Exec(
		"INSERT INTO data (chain, key, value) VALUES ($1, $2, $3) ON CONFLICT DO UPDATE SET value = $3",
		d.config.ChainConfig.Name,
		key,
		data,
	)
	if err != nil {
		d.logger.Error().Err(err).Str("key", key).Msg("Could not insert value")
		return err
	}

	return nil
}

func (d *Database) GetLastSnapshot() (*snapshotPkg.Info, error) {
	rawData, err := d.GetValueByKey("snapshot")
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not get snapshot")
		return nil, err
	}

	var snapshot snapshotPkg.Info
	if err := json.Unmarshal(rawData, &snapshot); err != nil {
		d.logger.Error().Err(err).Msg("Could not unmarshal snapshot")
		return nil, err
	}

	return &snapshot, nil
}

func (d *Database) SetSnapshot(snapshot *snapshotPkg.Info) error {
	rawData, err := json.Marshal(snapshot)
	if err != nil {
		d.logger.Error().Err(err).Msg("Could not marshal snapshot")
		return err
	}

	if err := d.SetValueByKey("snapshot", rawData); err != nil {
		d.logger.Error().Err(err).Msg("Could not save snapshot")
		return err
	}

	return nil
}
