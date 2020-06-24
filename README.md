# Overview #

This sample spins up a Redis server and Go work server in Docker Compose.

The work server first generates 100 example work tasks, each being a simple sleep job.
These are all added to Redis as Hashes, then the key of each task is added to a work queue.

The work server then spawns (cores - 1) workers which each take new tasks from the work queue and performs them.
Each worker publishes a finished message to a Redis channel when no more work can be retrieved from the queue.
A listener worker process received finish messages from a Redis subscription.

# Running #

`docker-compose up`

All output is printed to stdout. You should see that work is generated, picked up by workers in parallel, then each worker finishes before the program exits.

# Testing #

Unit testing would be done by subsituting function parameters, such as the Redis client, with a mock client that fulfils the same interface as the Redis client.
As most output of the program is via `fmt`, these calls would be substituted with calls to some logging interface that is also passed as a parameter. This can again be mocked.

# Shortcomings #

There is currently no writing back the result of the work tasks, or reporting of task completion.

If the task for some reason cannot be completed, it needs returning to the work queue - which is straightforward, but what if the worker encounters an error at this point?
It would be advisable to give work items a timeout and to have some other process checking for timed out tasks to return them to the work queue.
Redis' EXPIRE functionality can ensure that a task doesn't both become finished and returned to to the work queue.

Individual workers are currently geared to ditching the current task if a failur occurs and then continuing to work on new tasks rather than panic and abort their own process.
Work timeouts will help return jobs to the work queue.

# Scaling #

Work can be taken by workers in larger batches that just ones.
Work items could be bundled up into large work packages upon creation.
This would reduce the workers' overhead in communication with Redis as communication becomes less frequent.
Work packages could be "pinged" to let Redis know that the work package is still being worked on to prevent timeout. 

# Sequential Tasks #

Work is currently all performed in parallel but there might be cases where work items must be worked on one after the other.
Work items being generic hashes, they could contain entries pointing to tasks to attempt next and entries that require results from previous work items.
Workers returning results for work items can then look at "subsequent" work items to the current one and determine if all of those have fulfilled all of their dependencies.
If they have, they can be added to the work queue.

