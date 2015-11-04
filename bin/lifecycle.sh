#!/bin/bash

./sqs-broker --config=config-sample.json

####################################################################################################################################

# Catalog
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X GET "http://username:password@localhost:3000/v2/catalog"

####################################################################################################################################

# Provision SQS
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testsqs" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","organization_guid":"organization_id","space_guid":"space_id"}'

# Bind SQS
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testsqs/service_bindings/sqs-binding" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587"}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testsqs/service_bindings/sqs-binding-2" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587"}'

# Unbind SQS
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/testsqs/service_bindings/sqs-binding-2?service_id=c3c5ff4fd-3a27-4e66-a94d-f7d38558405d&plan_id=7105c0ec-5238-45d4-a08d-3d682fbb4587"
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/testsqs/service_bindings/sqs-binding?service_id=c3c5ff4fd-3a27-4e66-a94d-f7d38558405d&plan_id=7105c0ec-5238-45d4-a08d-3d682fbb4587"

# Update SQS
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testsqs" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","previous_values":{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","organization_guid":"organization_id","space_guid":"space_id"},"parameters":{"delay_seconds":"1"}}'

# Deprovision SQS
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/testsqs?service_id=3c5ff4fd-3a27-4e66-a94d-f7d38558405d&plan_id=7105c0ec-5238-45d4-a08d-3d682fbb4587"

####################################################################################################################################

# Provision Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testsqs" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","organization_guid":"organization_id","space_guid":"space_id","parameters":{"((("}}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testsqs" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"unknown","organization_guid":"organization_id","space_guid":"space_id"}'

# Update Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testsqs" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","previous_values":{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","organization_guid":"organization_id","space_guid":"space_id"},"parameters":{"((("}}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testsqs" -d '{"service_id":"unknown","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","previous_values":{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","organization_guid":"organization_id","space_guid":"space_id"},"parameters":{"delay_seconds":"1"}}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testsqs" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"unknown","previous_values":{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587","organization_guid":"organization_id","space_guid":"space_id"},"parameters":{"delay_seconds":"1"}}'

# Deprovision Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/unknown?service_id=3c5ff4fd-3a27-4e66-a94d-f7d38558405d&plan_id=7105c0ec-5238-45d4-a08d-3d682fbb4587"

# Bind Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testsqs/service_bindings/sqs-binding" -d '{"service_id":"unknown","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587"}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/unknown/service_bindings/sqs-binding" -d '{"service_id":"3c5ff4fd-3a27-4e66-a94d-f7d38558405d","plan_id":"7105c0ec-5238-45d4-a08d-3d682fbb4587"}'

# Unbind Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/testsqs/service_bindings/unknown?service_id=c3c5ff4fd-3a27-4e66-a94d-f7d38558405d&plan_id=7105c0ec-5238-45d4-a08d-3d682fbb4587"

# Last Operation Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X GET "http://username:password@localhost:3000/v2/service_instances/testsqs/last_operation"

####################################################################################################################################
