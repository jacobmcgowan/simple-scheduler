# Simple Scheduler

[![Build and Test Status](https://github.com/jacobmcgowan/simple-scheduler/actions/workflows/go-build-test.yml/badge.svg)](https://github.com/jacobmcgowan/simple-scheduler/actions/workflows/go-build-test.yml)
[![License: MIT](https://cdn.prod.website-files.com/5e0f1144930a8bc8aace526c/65dd9eb5aaca434fac4f1c34_License-MIT-blue.svg)](/LICENSE)

Simple service to schedule jobs to run at specific times or in intervals.

## Services
This monorepo contains multiple services for running and managing Simple
Scheduler. These can be found in the [services](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services) directory.

### Scheduler
This service is the worker service that schedules runs for jobs them. Currently
execution is handled via a message bus using a pub/sub model. When the scheduler
starts a run of a job it publishes a message to start the job. It is assumed
that a job worker service will consume this message then perform its task. The
job worker should publish status messages that the scheduler will consume to
update the run's status.

Currently, RabbitMQ is the only supported message bus but other options may be
added in future updates.

#### Adding support for alternative message bus services
To implement support for a different message bus
service, refer to the [MessageBus interface](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus/message-bus.go).
Add a directory for the new message bus option in [services/scheduler/message-bus](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus)
and add a constant for it to [MessageBusType](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus/message-bus-types/message-bus-types.go).
Lastly, update [RegisterMessageBus()](https://github.com/jacobmcgowan/simple-scheduler/blob/main/shared/resources/resources.go#L45)
to support the configuration option to use the new message bus service.

#### Running
The scheduler currently has the following dependencies:
- [MongoDB](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-community-with-docker/)
- [RabbitMQ](https://www.rabbitmq.com/docs/download)

Once the dependencies are running, you can run the service locally using
```bash
cd services/scheduler
go build .
./scheduler
```

or run in Docker using
```bash
docker build -f ./services/api/Dockerfile -t simple-scheduler/scheduler .
docker run -it --rm --name simple-scheduler -p 8080:8080 simple-scheduler/scheduler:latest
```

#### Settings
The scheduler supports the following settings set in the .env file:
| Environment Variable                          | Description                                                                                |
|-----------------------------------------------|--------------------------------------------------------------------------------------------|
| SIMPLE_SCHEDULER_MAX_JOBS                     | The maximum number of jobs the scheduler instance will manage. If 0, then no limit is set. |
| SIMPLE_SCHEDULER_DB_TYPE                      | The type of database to use. e.g. mongodb.                                                 |
| SIMPLE_SCHEDULER_MESSAGEBUS_TYPE              | The type of message bus to use. e.g. rabbitmq.                                             |
| SIMPLE_SCHEDULER_DB_CONNECTION_STRING         | The connection string of the database.                                                     |
| SIMPLE_SCHEDULER_DB_NAME                      | The name of the database to connect to.                                                    |
| SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING | The connection string of the message bus.                                                  |
| SIMPLE_SCHEDULER_CLEANUP_INTERVAL             | The interval in milliseconds to cleanup stuck runs.                                        |
| SIMPLE_SCHEDULER_CACHE_REFRESH_INTERVAL       | The interval in milliseconds to refresh the job cache.                                     |

### API
This service provides a RESTful API to manage jobs and runs.

#### Running
The API currently has the following dependencies:
- [MongoDB](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-community-with-docker/)

Once the dependencies are running, you can run the API locally using
```bash
cd services/api
go build .
./api
```

or run in Docker using
```bash
docker build -f ./services/api/Dockerfile -t simple-scheduler/api .
docker run -it --rm --name simple-scheduler-api -p 8080:8080 simple-scheduler/api:latest
```

### CLI
This application allows you to manage jobs and runs in a terminal.

#### Running
The CLI makes calls to the [API](#api) so ensure that is running.

To run locally use
```bash
cd services/cli
go build .
./cli --help
```
