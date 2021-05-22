Distributed Systems Class - Final Challenge
===========================================

This is the final challenge  for the Distributed Computing Class.

A strong recomendation is that you develop your solution the most simple, readable, scalable and plugable as possible. In the future you may reuse this code to
be integrated with  more services, so a well-defined design and implementation will make it easier to integrate new modules into your distributed application.

Distributed and Parallel Image Processing
-----------------------------------------

![architecture](images/architecture.png)

### Last Phase for the Final Challage
This is going to be the last phase of design and implementation.
On this phase you are working in all the components.
- API
- Controller
- Scheduler
- Worker

Your project will be divided on packages with very descriptive names where each system's component will be implemented.
Below you can see the details of each package and requirements for this final challenge:

### API
Below you will see the complete list of supported API endpoints.

  - All request must be token-based authenticated
  - Make sure that you have a proper error handling, the output format is up to you
  - Respect the parameters, output fields and http method
  - All responses must be in JSON format (except for the image's download)

#### Endpoints
| Endpoint                   | Authentication | Description                             | Supported Parameters      | JSON Response fields                                                                                                                                                                 | HTTP Method |
|----------------------------|----------------|-----------------------------------------|---------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|
| `/login`                   | User/Password  | Logins into DPIP                        |                           | `user`, `token`                                                                                                                                                                      | POST        |
| `/logout`                  | Token          | Logouts from DPIP                       |                           | `logout_message`                                                                                                                                                                     | DELETE      |
| `/status`                  | Token          | Provides overall system status          |                           | `system_name`, `server_time`, `active_workloads`(array)                                                                                                                              | GET         |
| `/workloads`               | Token          | Creates a new workload                  | `filter`, `workload_name` | `workload_id`, `filter` (`grayscale`, or `blur`), `workload_name`, `status` (`scheduling`, `running`, `completed`), `running_jobs` (integer), `filtered_images` (images - IDs array) | POST        |
| `/workloads/{workload_id}` | Token          | Gets details about an specific workload |                           | Same data as previous endpoint ^^                                                                                                                                                    | GET         |
| `/images`                  | Token          | Uploads an image                        | Image file, `workload_id` | `workload_id`, `image_id`, `type` (`orginal` or `filtered`)                                                                                                                          | POST        |
| `/images/{image_id}`       | Token          | Downloads an original or filtered image |                           |                                                                                                                                                                                      | GET         |


### Controller

The `controller` module has 2 main roles:

- A **message-passing** server that will be in communication where all new workers will subscribed
- A **datastore** service for system's overall status, workloads and workers information

Below some extra notes and considerations for the Controller

  - Consider reading the [Nanomsg documentation](https://nanomsg.org/gettingstarted/index.html) and [mangos package](https://github.com/nanomsg/mangos) in order to do the proper choice of Messaging Pattern for your system.
  - The Controller will keep record of activity and CPU and Memory resources utilization on its datastore service
  - The Controller will keep record of workloads information
  - For every `workload_name`, the controller is creating a directory in the `images` space
  - The `images` directory will serve for saving all procceded images that are coming from the workers
  - Image's name can be according to the ID you're assigning. It's your choice, just make sure that you have the proper information in the controller's database.


### Scheduler
  - Smart scheduling based on node utilization in terms of CPU, Memory and GPU availability
  - Scheduler is calling workers through RPC


### Worker

The `worker` will be a self-running component that will do the real work of filtering images.

- The `worker` will initially connect to the `controller` for:
  - Subscribing a a new `worker`
  - Once subscribed, receive information about the API (endpoint and token), this will be required by the `worker` in order to
    download and then upload images
- Worker will keep updated the controller with its CPU and Memory resources, alongside the number of running jobs
- You can run as many workers you want
- Scheduler will send jobs to the `worker` with RPC
- Startup command for the worker is:

```
cd worker/
export GO111MODULE=off
go run main.go --controller <host>:<port> --worker-name <worker_name> --tags <tag1>,<tag2>
```

- `image-store-endpoint` will be the API's endpoint
- `image-store-token` will serve for authenticating the Image Store API

## Documentation

- A detailed arquitecture must be written in the [architecture.md](architecture.md) file. Diagrams and charts can be included on this document.
- A detailed user guide must be written in the [user-guide.md](user-guide.md) file. This document explains how to install, configure and use your system.


Test Cases (from console)
-------------------------

**IMPORTANT**
This section will be updated soon, expect updates about the full system test suite.

- A [script](https://floobits.com/obedmr/dc-labs/file/final/stress_test.py) is provided to do an intensive end-to-end testing

"Game" Rules
------------

- This is 2-person team challenge, keep the focus on your work.
- You're free to use the internet for coding references.
- Any attempt of plagiarism will not be tolerated.


General Submission Instructions
-------------------------------
1. Make sure your local repository is in sync with the origin remote repository before anything.
2. Commit and Push your code to your personal repository (fork) and branch (first-partial).

3. Once you're done, follow common lab's sumission process. More details at: [Classify API](../../classify.md)
```
GITHUB_USER=<your_github_account> make submit

# Example:
GITHUB_USER=obedmr make submit
```

Grading Policy
--------------

The grading policy is quite simple, most falls in the test cases. Below the percentages table:

| Concept                                | %    |
|----------------------------------------|------|
| Code Style best practices              | 20%  |
| Test Cases (one for each API endpoint) | 60%  |
| Program meets with all requirements    | 20%  |
| TOTAL                                  | 100% |

Handy links
-----------
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Static File Server](https://github.com/gin-contrib/static)
- [Postman](https://www.postman.com/)
- [Video: Basics of Using Postman](https://youtu.be/t5n07Ybz7yI)
- [Advanced REST client for Chrome browser](https://chrome.google.com/webstore/detail/advanced-rest-client/hgmloofddffdnphfgcellkdfbfbjeloo?hl=es-419)