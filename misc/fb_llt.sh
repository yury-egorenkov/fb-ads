#!/bin/sh

#
# Get token for page here:
#
#   https://developers.facebook.com/tools/explorer/
#
# Exchange short-living Facebook access token to long-living (60d) Facebook access token.
#

ERR_MSG="missing keys"

. "$(dirname "$0")/util" &&
req 'FACEBOOK_PAGE_VER' "$ERR_MSG" &&
req 'FACEBOOK_PAGE_APP_ID' "$ERR_MSG" &&
req 'FACEBOOK_PAGE_APP_SECRET' "$ERR_MSG" &&
req 'FACEBOOK_PAGE_ACCESS_TOKEN' "$ERR_MSG" &&

curl -i -X GET "https://graph.facebook.com/v${FACEBOOK_PAGE_VER}/oauth/access_token?grant_type=fb_exchange_token&client_id=${FACEBOOK_PAGE_APP_ID}&client_secret=${FACEBOOK_PAGE_APP_SECRET}&fb_exchange_token=${FACEBOOK_PAGE_ACCESS_TOKEN}"
