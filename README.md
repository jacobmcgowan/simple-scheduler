# Simple Scheduler
Simple service to schedule jobs to run at specific times or in intervals

## Services
This monorepo contains multiple services for running and managing Simple
Scheduler. These can be found in the [services](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services) directory.

### Scheduler
This service is the schedules runs for jobs and executes them. Currently
execution is handled via a message bus using a pub/sub model. When the scheduler
starts a run of a job it publishes a message to start the job. It is assumed
that a job worker service will consume this message then perform its task. The
job worker should publish status messages that the scheduler will consume to
update the run's status.

Currently, RabbitMQ is the only supported message bus but other options may be
added in future updates.

#### Adding support for alternative message bus services
To implement support for a different message bus
service, refer to the [MessageBus interface](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus/message-bus.go). Add a
directory for the new message bus option in [services/scheduler/message-bus](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus) and add a constant for it to [MessageBusType](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus/message-bus-types/message-bus-types.go). Lastly, update [RegisterMessageBus()](https://github.com/jacobmcgowan/simple-scheduler/blob/main/shared/resources/resources.go#L45) to
support the configuration option to use the new message bus service.

#### Running
The scheduler currently has the following dependencies:
- [MongoDB](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-community-with-docker/)
- [RabbitMQ](https://www.rabbitmq.com/docs/download)

Once the dependencies are running, simply run
```bash
cd services/scheduler
go build .
./scheduler
```

or run in Docker
```bash
docker build -f ./services/api/Dockerfile -t simple-scheduler/api .
docker run -it --rm --name simple-scheduler-api -p 8080:8080 simple-scheduler/api:latest
```

### API
This service provides a RESTful API to get details about and manage jobs.

#### Running
The API currently has the following dependencies:
- [MongoDB](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-community-with-docker/)

Once the dependencies are running, simply run
```bash
cd services/api
go build .
./api
```

### CLI
This application allows you to view and manage jobs and runs in a terminal.

#### Running
The CLI makes calls to the [API](#api) so ensure that is running.

To run use
```bash
cd services/cli
go build .
./cli --help
```
