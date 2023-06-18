#!/usr/bin/env fish

echo "Starting Server..."

trap "echo 'Stopping Server...'; kill (cat .server_pid) && rm .server_pid; exit 0" SIGINT SIGTERM

while true
    
    if [ -f .server_pid ]
        kill (cat .server_pid)
    end

    cargo run --bin yoake_server -- -c config-dev.yaml & echo $last_pid > .server_pid
    inotifywait -e modify -e move -e create -e delete -r src \
        && kill (cat .server_pid)
end