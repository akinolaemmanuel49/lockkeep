# use PRIMARY
docker cp .\seed\users.json lockkeep-backend-mongo3-1:/tmp/users.json
docker cp .\seed\vault.json lockkeep-backend-mongo3-1:/tmp/vault.json


docker exec lockkeep-backend-mongo3-1 mongoimport --username root --password password --authenticationDatabase admin --db lockkeep --collection users --jsonArray --file /tmp/users.json
docker exec lockkeep-backend-mongo3-1 mongoimport --username root --password password --authenticationDatabase admin --db lockkeep --collection vault --jsonArray --file /tmp/vault.json