#!/usr/bin/env bash

# run the server
umask 0002
cd bin; ./illiadupload  -dir $ILLIAD_UPLOAD_PATH -jwtkey $JWT_KEY

# return the status
exit $?
