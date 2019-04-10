#!/bin/bash

aws s3 cp publish/trackman* s3://downloads.cloud66.com/trackman --acl public-read
