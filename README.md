Distributed Systems Class - Final Challenge
===========================================

This is the final challenge  for the Distributed Computing Class. This is the integration of 3 phases.

A strong recomendation is that you develop your solution the most simple, readable, scalable and plugable as possible. In the future you may reuse this code to
be integrated with  more services, so a well-defined design and implementation will make it easier to integrate new modules into your distributed application.

Distributed and Parallel Image Processing
-----------------------------------------

![architecture](architecture.png)

### Last Phase for the Final Challage
This is going to be the last phase of design and implementation.
On this phase you are working in all the components.
- API
- Controller
- Scheduler
- Worker

Your project will be divided on packages with very descriptive names where each system's component will be implemented.
Below you can see the details of each package and requirements for this final challenge:

- `api/`
  - All request must be token-based authenticated
  - **Endpoint:** `/results/<workload_id>` - Will serve as a static file server for all processed images for the specified workload id
  - **Endpoint:** `workloads/filter` - This endpoint will trigger an image filtering. This will be an end-to-end call from `api` to `worker`.
  - **Endpoint:** `/download` - Will be used by workers to download images that will be filtered
  - **Modify Endpoint:** `/upload` - Now, it will be only used by workers to store filtered images
  - For last 2 endpoints, workers will authenticate with workers-specific tokens, user's tokens will not work for these endpoints.

- `controller/`
  - Controller will keep record of activity and CPU,Memory and GPU resources utilization on its data store mechanism
  - Controller will keep record of the workloads information
  - For every workload id, the controller is creating a results directory
  - The results directory will server for saving all procceded images that are coming from the workers
  - Image's name will be renamed in a consecutive order as they were arriving in time. Below an example on how **results** directory should look:
  ```
	results/my-filters/
		1.png
		2.png
		3.png
		4.png
		5.png
	results/video-frames/
		1.jpg
		2.jpg
		3.jpg
  ```

- `scheduler/`
  - Smart scheduling based on node utilization in terms of CPU, Memory and GPU availability
  - Scheduler is calling workers through RPC

- `worker/`
  - Standalone component with initial `test` RPC function.
  - Worker's command line will be as follows:
    - ```
      ./worker --controller <host>:<port> --node-name <node_name> --tags <tag1>,<tag2> \
      	       --image-store-endpoint <host>:<port> --image-store-token <auth-token>
      ```
  - `image-store-endpoint` will be the API's endpoint
  - `image-store-token` will serve for authenticating the Image Store API

**Documentation**
- A detailed arquitecture document will be required for this initial phase in the [architecture.md](architecture.md) file. Diagrams and charts can be included on this document.
- A detailed user guide must be written in the [user-guide.md](user-guide.md) file. This document explains how to install, configure and use your system.


Test Cases (from console)
-------------------------
- **Execute Filter Workload**
```
$ curl -F 'data=@path/to/local/image.png' -d 'workload-id=my-filters&filter=grayscale' -H "Authorization: Bearer <ACCESS_TOKEN>" http://localhost:8080/workloads/filter
{
	"Workload ID": "my-filters",
	"Filter": "grayscale",
	"Job ID": 1,
	"Status": "Scheduling",
	"Results: "http://localhost:8080/results/my-filters/"
}
```

- **WORKERS API calls**
  - http://localhost:8080/upload
    - Request will contain `workload_id` and `image`, authenticated with the `worker-token`
  - http://localhost:8080/download
    - Request will contain `workload_id` and `image_id`, authenticated with the `worker-token`

- **Results Endpoint**
  - http://localhost:8080/results/<workload_id>

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