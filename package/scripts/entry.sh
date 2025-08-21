#!/usr/bin/env bash

# run the server
umask 0002
cd bin; ./illiadupload  -dir ${ILLIAD_UPLOAD_PATH}

# return the status
exit $?

#
# end of file
#
