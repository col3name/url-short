docker run -v ./data/migrations:/migrations  --network host migrate/migrate  -path=/migrations/ -database postgresql://url-short:1234@localhost:5432/url-short?sslmode=disable %1 %2