[![Build Status](https://travis-ci.org/HZ89/djob.svg?branch=master)](https://travis-ci.org/HZ89/djob)
[![Go Report Card](https://goreportcard.com/badge/github.com/hz89/djob)](https://goreportcard.com/report/github.com/hz89/djob)

# Summary description
djob supports cron syntax (second-level supported) for scheduled and one-time job. Select the
server to execute the job through the user-defined tags. Job is randomly assigned to the server in the 
same region for scheduling. If any one server failed, the job it managed will automatically take over 
by another.
# How to install
* git clone
* cd djob
* glide install
* rm -rf vendor/github.com/hashicorp/serf/vendor
* go build
# API DOC
https://documenter.getpostman.com/view/2679557/djob-api/77o2fVz
