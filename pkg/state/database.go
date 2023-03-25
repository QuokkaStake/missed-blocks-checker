package state

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	configPkg "main/pkg/config"
	"main/pkg/types"
	migrations "main/sql"
	"time"
)

type Database struct {
	Logger zerolog.Logger
	Config *configPkg.Config
	Client *sql.DB
}

func NewDatabase(logger *zerolog.Logger, config *configPkg.Config) *Database {
	return &Database{
		Logger: logger.With().Str("component", "state_manager").Logger(),
		Config: config,
	}
}

func (d *Database) Init() {
	db, err := sql.Open("sqlite3", d.Config.DatabaseConfig.Path)

	if err != nil {
		d.Logger.Fatal().Err(err).Msg("Could not open sqlite database")
	}

	var version string
	err = db.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		d.Logger.Fatal().Err(err).Msg("Could not query sqlite database")
	}

	d.Logger.Info().
		Str("version", version).
		Str("path", d.Config.DatabaseConfig.Path).
		Msg("sqlite database connected")

	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		d.Logger.Fatal().Err(err).Msg("Could not get migrations folder path")
	}

	for _, entry := range entries {
		d.Logger.Info().
			Str("name", entry.Name()).
			Msg("Applying sqlite migration")

		content, err := migrations.FS.ReadFile(entry.Name())
		if err != nil {
			d.Logger.Fatal().
				Str("name", entry.Name()).
				Err(err).
				Msg("Could not read migration content")
		}

		statement, err := db.Prepare(string(content))
		if err != nil {
			d.Logger.Fatal().
				Str("name", entry.Name()).
				Err(err).
				Msg("Could not prepare migration")
		}
		if _, err := statement.Exec(); err != nil {
			d.Logger.Fatal().
				Str("name", entry.Name()).
				Err(err).
				Msg("Could not execute migration")
		}
	}

	d.Client = db
}

func (d *Database) InsertBlock(block *types.Block) error {
	ctx := context.Background()
	tx, err := d.Client.BeginTx(ctx, nil)
	if err != nil {
		d.Logger.Error().Err(err).Msg("Could not create a transaction for saving a block")
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO blocks (height, time, proposer) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		block.Height,
		block.Time.Unix(),
		block.Proposer,
	)
	if err != nil {
		d.Logger.Error().Err(err).Msg("Error saving block")
		if err = tx.Rollback(); err != nil {
			d.Logger.Error().Err(err).Msg("Error rolling back transaction")
			return err
		}

		return err
	}

	for key, signature := range block.Signatures {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO signatures (height, validator_address, signature) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			block.Height,
			key,
			signature,
		)
		if err != nil {
			d.Logger.Error().Err(err).Msg("Error saving signature")
			if err = tx.Rollback(); err != nil {
				d.Logger.Error().Err(err).Msg("Error rolling back transaction")
				return err
			}

			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		d.Logger.Error().Err(err).Msg("Could not commit a transaction")
		return err
	}

	return nil
}

func (d *Database) GetAllBlocks() (map[int64]*types.Block, error) {
	blocks := map[int64]*types.Block{}

	// Getting blocks
	blocksRows, err := d.Client.Query("SELECT height, time, proposer FROM blocks")
	if err != nil {
		d.Logger.Error().Err(err).Msg("Error getting all blocks")
		return blocks, err
	}

	for blocksRows.Next() {
		var (
			blockHeight   int64
			blockTime     int64
			blockProposer string
		)

		err = blocksRows.Scan(&blockHeight, &blockTime, &blockProposer)
		if err != nil {
			d.Logger.Error().Err(err).Msg("Error fetching block data")
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
	signaturesRows, err := d.Client.Query("SELECT height, validator_address, signature FROM signatures")
	if err != nil {
		d.Logger.Error().Err(err).Msg("Error getting all blocks")
		return blocks, err
	}

	for signaturesRows.Next() {
		var (
			signatureHeight int64
			validatorAddr   []byte
			signature       int32
		)

		err = signaturesRows.Scan(&signatureHeight, &validatorAddr, &signature)
		if err != nil {
			d.Logger.Error().Err(err).Msg("Error fetching signature data")
			return blocks, err
		}

		_, ok := blocks[signatureHeight]
		if !ok {
			d.Logger.Fatal().
				Int64("height", signatureHeight).
				Msg("Got signature for block we do not have, which should never happen.")
		}

		blocks[signatureHeight].Signatures[string(validatorAddr)] = signature
	}

	return blocks, nil
}

func (d *Database) TrimBlocksBefore(height int64) error {
	ctx := context.Background()
	tx, err := d.Client.BeginTx(ctx, nil)
	if err != nil {
		d.Logger.Error().Err(err).Msg("Could not create a transaction for trimming blocks")
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM blocks WHERE height <= $1", height)
	if err != nil {
		d.Logger.Error().Err(err).Msg("Error trimming blocks")
		if err = tx.Rollback(); err != nil {
			d.Logger.Error().Err(err).Msg("Error rolling back transaction")
			return err
		}

		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM signatures WHERE height <= $1", height)
	if err != nil {
		d.Logger.Error().Err(err).Msg("Error trimming signatures")
		if err = tx.Rollback(); err != nil {
			d.Logger.Error().Err(err).Msg("Error rolling back transaction")
			return err
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		d.Logger.Error().Err(err).Msg("Could not commit a transaction")
		return err
	}

	return nil
}