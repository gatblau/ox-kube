#!/usr/bin/env bash
oc apply -f namespace.yml
oc apply -f quota.yml
oc apply -f pod.yml
oc apply -f service.yml
