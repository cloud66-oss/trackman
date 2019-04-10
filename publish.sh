#!/bin/bash

aws s3 cp build s3://downloads.cloud66.com/trackman --acl public-read --recursive
