#!/bin/bash

# A script to generate varied log lines to showcase loglit features
while true; do
  # 1. Standard App Log (INFO/WARN with Dates, UUIDs, Numbers)
  echo "$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ") [INFO] User login successful | user_id=550e8400-e29b-41d4-a716-446655440000 | ip=192.168.1.42 | duration=145ms"
  sleep 0.2
  # 2. Database Log (DEBUG with SQL, Strings, Booleans)
  echo "$(date +"%Y/%m/%d %H:%M:%S") [DEBUG] Executing query: \"SELECT * FROM users WHERE active = true AND role != null\" | rows=120"
  sleep 0.2
  # 3. Error Log (ERROR with Path, Exception, Hex numbers)
  echo "$(date +"%b %d %H:%M:%S") [ERROR] Connection refused to service at https://api.internal.svc:8080 | error_code=0x5F3 | retries=3"
  echo "    Caused by: java.net.ConnectException: Connection timed out"
  echo "    at com.example.service.Network.connect(Network.java:45)"
  sleep 0.3
  # 4. Critical/Fatal (FATAL/CRITICAL with MAC Addr, Memory addresses)
  echo "[$(date +%T)] [FATAL] Hardware fault detected on device aa:bb:cc:11:22:33 | memory_addr=0x7fff5fbff7c0 | temp=95.5 C"
  sleep 0.5
  # 5. Verbose Trace (TRACE with complex floats, binary, JSON-like data)
  echo "T: $(date +%s) [TRACE] Sensor data dump: { \"v1\": 12.345, \"v2\": 1.5e-3, \"flags\": 0b10110011, \"valid\": false }"
  sleep 0.2
  # 6. Separator to show structure
  echo "----------------------------------------------------------------"
  sleep 0.1
done
