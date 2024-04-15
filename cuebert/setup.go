package main

import (
	"os"
	"strings"
	"sync/atomic"

	"github.com/Masterminds/squirrel"
	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/bot/lifecycle"
	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/cuebert/task/client"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/idp"
	idpclient "github.com/johnmikee/cuebert/idp/client"
	"github.com/johnmikee/cuebert/mdm"
	mdmclient "github.com/johnmikee/cuebert/mdm/client"
)

func setup() *Cuebert {
	cb, _ := loadEnv()

	conn, cfg := dbConn(cb, false)
	cb.db = conn
	cb.dbConfig = cfg

	cycle := lifecycle.New()
	cb.statusChan = make(chan handlers.StatusMessage)
	cb.statusHandler = &handlers.StatusHandler{}
	cb.lifecyle = cycle

	// configure lifecycle with atomic operations
	cb.lifecyle.SetRunningFunc(
		func() bool {
			return atomic.LoadInt32(&cb.running) == 1
		},
		func(status bool) {
			val := int32(0)
			if status {
				val = 1
			}
			atomic.StoreInt32(&cb.running, val)
		},
	)
	cb.bot = bot.New(
		&bot.Config{
			Flags:         cb.flags,
			Config:        cb.config,
			Log:           cb.log,
			LifeCycle:     cycle,
			StatusChan:    cb.statusChan,
			StatusHandler: cb.statusHandler,
		},
	)
	idp := idpclient.New(&idpclient.IDP{
		IDP: idp.IDP(cb.flags.IDP),
		Config: idp.Config{
			Domain:        cb.config.Subdomain,
			URL:           cb.config.IDPURL,
			Token:         cb.config.IDPToken,
			Log:           cb.log,
			IDPBlockGroup: cb.config.IDPBlockGroup,
		},
	})

	if cb.flags.AuthUsersFromIDP {
		oid, err := idp.GetAdminGroup(cb.config.AdminGroupID)
		if err != nil {
			cb.log.Info().AnErr("error", err).Msg("could not get admin group")
			os.Exit(3)
		}
		for _, e := range oid {
			cb.log.Trace().Str("user", e).Msg("adding user to authorized users")
			sid, err := cb.bot.Bot().SlackClient().GetUserByEmail(e)
			if err != nil {
				cb.log.Debug().Str("user", e).Msg("could not get user id")
				continue
			}
			cb.flags.AuthUsers = append(cb.flags.AuthUsers, sid.ID)
		}
	} else if cb.flags.AuthorizedUsers != "" {
		userSlice := strings.Split(cb.flags.AuthorizedUsers, ",")
		cb.flags.AuthUsers = userSlice
	}

	cb.task = client.New(
		&client.TaskConfig{
			Task: task.Task{
				Option:         task.Option(cb.flags.Task),
				Implementation: implementation.Item(cb.flags.TaskImplementation),
			},
			Config: &task.Config{
				DBConfig:      cfg,
				LifeCycle:     cycle,
				Log:           cb.log,
				DB:            conn,
				Pool:          conn.GetConnection(),
				Bot:           cb.bot,
				StatusHandler: cb.statusHandler,
				SlackClient:   cb.bot.Bot().SlackClient(),
				Shared:        &shared.Shared{},
				IDP:           idp,
				MDM: mdmclient.New(
					&mdmclient.MDM{
						MDM: mdm.MDM(cb.flags.MDM),
						Config: mdm.Config{
							Domain: cb.config.Subdomain,
							MDM:    mdm.MDM(cb.flags.MDM),
							URL:    cb.config.MDMURL,
							Token:  cb.config.MDMKey,
							Client: nil,
							Log:    cb.log,
						},
					},
				),
				Flags:  cb.flags,
				Config: cb.config,
			},
		})

	cb.bot.Update(bot.WithLifeCycle(cb.task), bot.WithFlags(cb.flags), bot.WithInteractive(cb.task))
	return cb
}

func dbConn(cb *Cuebert, test bool) (db.Client, *config.Config) {
	dbconn := &config.Config{
		Connection: config.Connection{
			Host:     cb.config.DBAddress,
			Name:     cb.config.DBName,
			Password: cb.config.DBPass,
			Port:     cb.config.DBPort,
			User:     cb.config.DBUser,
			SSLMode:  "disable",
		},
		Type:     config.Postgres,
		TaskName: cb.flags.TaskImplementation,
		Log:      &cb.log,
		ST:       squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	db, err := db.New(dbconn)
	if err != nil {
		cb.log.Info().AnErr("connecting to db", err).Send()
		os.Exit(3)
	}

	if test {
		err = db.Ping()
		if err != nil {
			cb.log.Info().AnErr("connecting to db", err).Send()
			os.Exit(3)
		}
		cb.log.Info().Msg("db connection successful")
		os.Exit(0)
	}

	dbconn.DB.DB = db.GetConnection()
	dbconn.DB.Pool = db.GetPool()

	return db, dbconn
}
