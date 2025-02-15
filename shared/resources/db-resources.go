package resources

import (
	"fmt"
	"net/url"
	"os"

	dbTypes "github.com/jacobmcgowan/simple-scheduler/shared/data-access/db-types"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	mongoRepos "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/mongo"
	envVars "github.com/jacobmcgowan/simple-scheduler/shared/resources/env-vars"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DbEnv struct {
	Type             string
	ConnectionString string
	Name             string
}

type DbResources struct {
	Name        string
	Context     repositories.DbContext
	ManagerRepo repositories.ManagerRepository
	JobRepo     repositories.JobRepository
	RunRepo     repositories.RunRepository
}

func LoadDbEnv() DbEnv {
	return DbEnv{
		Type:             os.Getenv(envVars.DbType),
		ConnectionString: os.Getenv(envVars.DbConnectionString),
		Name:             os.Getenv(envVars.DbName),
	}
}

func RegisterRepos(env DbEnv) (DbResources, error) {
	conStrUrl, err := url.Parse(env.ConnectionString)
	if err != nil {
		dbResources := DbResources{
			Name:    "",
			Context: nil,
			JobRepo: nil,
			RunRepo: nil,
		}
		return dbResources, fmt.Errorf("connection string invalid: %s", err)
	}

	switch env.Type {
	case string(dbTypes.MongoDb):
		dbCtx := mongoRepos.MongoDbContext{
			DbName:  env.Name,
			Options: *options.Client().ApplyURI(env.ConnectionString),
		}
		mngrRepo := mongoRepos.MongoManagerRepository{
			DbContext: &dbCtx,
		}
		jobRepo := mongoRepos.MongoJobRepository{
			DbContext: &dbCtx,
		}
		runRepo := mongoRepos.MongoRunRepository{
			DbContext: &dbCtx,
		}

		dbResources := DbResources{
			Name:        env.Name + "@" + conStrUrl.Host,
			Context:     &dbCtx,
			ManagerRepo: mngrRepo,
			JobRepo:     jobRepo,
			RunRepo:     runRepo,
		}
		return dbResources, nil
	default:
		dbResources := DbResources{
			Name:    conStrUrl.Host,
			Context: nil,
			JobRepo: nil,
			RunRepo: nil,
		}
		return dbResources, fmt.Errorf("DB type %s not supported", env.Type)
	}
}
