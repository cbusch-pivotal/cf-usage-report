#!/bin/bash
set -ex

# audit user information
AUDIT_USER="apptio-pcf-auditor"
AUDIT_PWD="Appt10intX17"
AUDIT_EMAIL="apptiopcfauditor@company.com"

# set target environment in which to create users
#uaac target uaa.system.<DOMAIN.COM> --skip-ssl-validation
uaac target uaa.system.mypcf.net --skip-ssl-validation

# Note: insert token after '-s' from Elastic Runtime tile -> Credentials tab -> UAA / Admin Client Credentials
#uaac token client get admin -s <UAA ADMIN CLIENT PASSWORD>
uaac token client get admin -s Jt5YRvF8VQWH2laqo_W159gsX--KveZ8
#uaac token client get usage_service -s U_AIZXItMXurjHvMT0WIl52X2365WyqZ

# create audit user
uaac user add $AUDIT_USER -p $AUDIT_PWD --emails $AUDIT_EMAIL
uaac member add cloud_controller.admin $AUDIT_USER
#uaac member add cloud_controller.global_auditor $AUDIT_USER
#uaac member add usage_service.audit $AUDIT_USER
#uaac member add cloud_controller.admin_read_only $AUDIT_USER
