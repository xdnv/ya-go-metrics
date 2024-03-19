rem server.exe -f .\log.json -i 0
server.exe -f .\log.json -d "host=BAD port=5432 user=postgres password=admin dbname=postgres sslmode=disable"
