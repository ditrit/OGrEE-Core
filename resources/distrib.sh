#!/bin/bash

export CRBIN=/usr/local/bin
export COCKROACH=$CRBIN/cockroach

export CRDATA=/var/lib/cockroach
export CERTDIR=$CRDATA/certs
export SAFEDIR=$CRDATA/safe
#DSNDIR=$(echo $CERTDIR | sed -e 's:/:%2F:g')

##
## 3-nodes-cluster / spread on 2 sites 
##
## INTERNAL @home + external @ orness ---> from vM:netbox
#export CRNODES_CR="cockroach.orness.local:26257 cockroach2.orness.local:26258 cockroach.orness.net:26257"
#export CRNODES_SSH="cockroach.orness.local:22 cockroach2.orness.local:22 cockroach.orness.net:2225"

## EXTERNAL for all -> from ziad@home
#export CRNODES_CR="www.chibois.net:26257 www.chibois.net:26258 cockroach.orness.net:26257"
#export CRNODES_SSH="www.chibois.net:2225 www.chibois.net:2226  cockroach.orness.net:2225"

## INTERNAL @home + @orness (when 2ite-to-site VPN is working) --> from VPN/VM:netbox
#export CRNODES_CR="cockroach.orness.local:26257 cockroach2.orness.local:26258 cockroach3.orness.local:26257"
#export CRNODES_SSH="cockroach.orness.local:22 cockroach2.orness.local:22 cockroach3.orness.net:22"

##
## 2-nodes-cluster / one site
##
export CRNODES_CR="cockroach.orness.local:26257 cockroach2.orness.local:26258"
export CRNODES_SSH="cockroach.orness.local:22 cockroach2.orness.local:22"

export CRMEMBERS=$(echo $CRNODES_CR |tr " " ",")
export FIRSTNODE=$(echo $CRNODES_CR|awk '{print $1}')

#
# title 
#
title () {

   len=${#1}
   ch='='
   printf '%*s\n' "$len" | tr ' ' "$ch"
   echo $1
   printf '%*s\n' "$len" | tr ' ' "$ch"
}

#
# check_nodes
#
# check DNS then PING then SSH
#
check_nodes() { 

   title "Checking nodes..."
   for i in $CRNODES_SSH
   do
      h=$(echo $i | awk -F\: '{print $1}')
      p=$(echo $i | awk -F\: '{print $NF}')
      echo "   Dealing with node : $h ... "
      echo -n "      Checking DNS/host ... "
      host $h >/dev/null
      if [[ $? -eq 0 ]]
      then 
         echo "OK"
         echo -n "      Pinging ..."
         ping -c1 $h > /dev/null
         if [[ $? -eq 0 ]]
         then 
            echo "OK"
         else 
            echo "KO"
         fi
         echo -n "      Ssh into node ...on port ..."
         ssh -p $p root@$h "date" > /dev/null
         if [[ $? -eq 0 ]]; then echo "OK"; else echo "KO"; fi
      else echo "KO"
      fi
   done
}

#
# cert_create_ca
#
cert_create_ca() {

   title "Cockroach CA creation"
   rm -rf $SAFEDIR $CERTDIR
   mkdir -p $SAFEDIR $CERTDIR
   $DEBUG $COCKROACH cert create-ca \
   --certs-dir=$CERTDIR \
   --ca-key=$SAFEDIR/ca.key \
   --overwrite
}

#
# cert_create_node
#
# https://www.cockroachlabs.com/docs/v20.2/deploy-cockroachdb-on-premises
#
cert_create_node() {

   if [ $# -eq 0 ]
   then
      h=$(hostname)
   else 
      h=$1
   fi
   title "   Cockroach node certificate creation for $h"
   rm -f $CERTDIR/node.crt $CERTDIR/node.key
   $DEBUG $COCKROACH cert create-node \
   localhost \
   $h \
   --certs-dir=$CERTDIR \
   --ca-key=$SAFEDIR/ca.key \
   --overwrite
}

#
# cert_create_nodes
#
cert_create_nodes() { 

   title "Creating node certificates ..."
   for i in $CRNODES_CR
   do
      h=$(echo $i | awk -F\: '{print $1}')
      cert_create_node $h
      cert_list_cert
      ssh root@$h "mkdir -p $CERTDIR"
      scp -r $CERTDIR $h:$CRDATA
   done
}

#
# cert_create_client
#
cert_create_client() {

   if [ $# -eq 0 ]
   then
      c=root
   else
      c=$1   
   fi
   title "Cockroach client certificate creation for $c"
   $DEBUG $COCKROACH cert create-client \
   $c \
   --certs-dir=$CERTDIR \
   --ca-key=$SAFEDIR/ca.key
}

#
# cert_list_cert
#
cert_list_cert() {
  title "List certificates"
  $COCKROACH cert list --certs-dir=$CERTDIR
}

#
# securecockroach_service
#
securecockroach_service() { 
   wget -q https://raw.githubusercontent.com/cockroachdb/docs/master/_includes/v20.2/prod-deployment/securecockroachdb.service
   sed -i -e 's!/var/lib/cockroach!#CRDATA#!g' -e 's!--advertise-addr=<node1 address>!--advertise-addr=#NODEADDR#!g' -e 's!--join=<node1 address>,<node2 address>,<node3 address>!--join=#CRNODES#!g' securecockroachdb.service
}

#
# deploy_nodes
#
deploy_nodes() { 

   title "Deploying cockroach on cluster nodes ..."
   for i in $CRNODES_SSH
   do 
      # $i=cockroach.orness.local:22
      h=$(echo $i | awk -F\: '{print $1}')
      p=$(echo $i | awk -F\: '{print $NF}')
      echo "   on $h node ..."
      echo "      copying binary to $CRBIN ..."
      scp $COCKROACH $h:$CRBIN
      echo "      creating cockroach user ..."
      ssh $h '$(getent passwd cockroach >/dev/null 2>&1) || useradd cockroach'
      ssh $h "chown -R cockroach.cockroach $CRDATA"
      echo "      deploying custom cockroach.service file  ..."
      sed -e "s!#CRDATA#!$CRDATA!g" -e "s!#NODEADDR#!$i!g" -e "s!#CRNODES#!$CRMEMBERS!g" securecockroachdb.service > securecockroachdb.service.$h
echo scp -P $p securecockroachdb.service.$h $h:/etc/systemd/system/securecockroachdb.service
      $DEBUG scp -P $p securecockroachdb.service.$h $h:/etc/systemd/system/securecockroachdb.service
   done
}

#
# start_nodes
#
start_nodes() { 

   title "Starting cockroach on cluster nodes ..."
   for i in $CRNODES_SSH
   do
      h=$(echo $i | awk -F\: '{print $1}')
      echo "   on $h node ..."
      ssh $h "systemctl start securecockroachdb"
      ssh $h "systemctl status securecockroachdb"
   done
}

#
# init_cluster
#
init_cluster() {
   echo "Cockroach cluster initialisation"
   $COCKROACH init \
		--certs-dir=$CERTDIR \
		--host=$FIRSTNODE 
}

#
# db_user_creation
#
db_user_creation() {

   title "OGrEE database and user creation"
   $COCKROACH sql 	\
	   --certs-dir=$CERTDIR \
		--host=$FIRSTNODE <<EOF
CREATE DATABASE IF NOT EXISTS ogree;
CREATE USER IF NOT EXISTS ogree WITH PASSWORD '0gr33!';
GRANT ALL ON DATABASE ogree TO ogree;
EOF
}

#
# clean_all
#
clean_all() {
   
   title "Stopping and cleaning cockroach on cluster nodes ..."
   for i in $CRNODES_SSH
   do
      h=$(echo $i | awk -F\: '{print $1}')
      echo "   on $h node ..."
      ssh $h "systemctl stop securecockroachdb"
      ssh $h "rm -f /etc/systemd/system/securecockroachdb.service"
      ssh $h "killall cockroach"
      ssh $h "rm -rf $CRDATA"
      ssh $h "rm -f $COCKROACH"
   done
}


#
# MAIN
#

check_nodes
cert_create_ca
cert_list_cert
cert_create_nodes
cert_create_client root
#cert_create_client ziad
cert_list_cert
securecockroach_service
deploy_nodes
start_nodes
init_cluster
db_user_creation
#clean_all