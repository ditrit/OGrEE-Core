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