#!/bin/sh

set -e

NATS_SERVICE_NAME=stan

CONFIG_FILE="/etc/gnatsd.conf"

if [ -n "$NATS_SERVICE_NAME" ]; then

    # get current IPs
    CURRENT_IP=`hostname -i | awk '{print $1}'`
    # fetch IPs from service members excepted CURRENT_IP
    CLUSTER_IPS=`nslookup $NATS_SERVICE_NAME | grep Address | awk '{print $3}' | awk 'NF' | sed "/$CURRENT_IP/d"`

    # we check if we are the only/first node live
    if [ `echo "$CLUSTER_IPS" | wc -l` = 0 ]; then
        CLUSTER_ADDRESS="[]"
    else
        # format CLUSTER_IPS for use in nats
        CLUSTER_MEMBERS=`echo "$CLUSTER_IPS" | sed "s|.*|nats-route://@&:6222|" | head -c -1 | tr '\n' ','`
        # we set cluster address with found service members
        CLUSTER_ADDRESS="[$CLUSTER_MEMBERS]"
    fi

    sed -i "s|routes = .*|routes = $CLUSTER_ADDRESS|" $CONFIG_FILE
fi

exec "$@"