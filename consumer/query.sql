-- name: AddSensorData :exec
INSERT INTO living_room (
  time, temperature, humidity
) VALUES (
  $1, $2, $3
);