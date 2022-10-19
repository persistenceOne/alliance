#!/bin/bash

printf "#1) Submit proposal to creaet ulunax Alliance...\n\n"
allianced tx gov submit-proposal gov.json --from=demowallet1 --home ./data/alliance --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

sleep 2
PROPOSAL_ID=$(allianced query gov proposals --count-total --output json | jq .pagination.total -r)

printf "\n#2) Deposit funds to the proposal...\n\n"
allianced tx gov deposit $PROPOSAL_ID 10000000stake --from=demowallet1 --home ./data/alliance --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

printf "\n#3) Vote to pass the proposal...\n\n"
allianced tx gov vote $PROPOSAL_ID yes --from=val1 --home ./data/alliance --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

printf "\n#4) Query proposals...\n\n"
allianced query gov proposal $PROPOSAL_ID --home ./data/alliance

printf "\n#5) Query alliances...\n\n"
allianced query alliance alliances --home ./data/alliance

printf "\n#6) Waiting for gov proposal to pass...\n\n"
sleep 5

printf "\n#7) Query alliances after proposal passed...\n\n"
allianced query alliance alliances --home ./data/alliance