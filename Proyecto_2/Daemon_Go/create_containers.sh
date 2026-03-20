#!/bin/bash

for i in {1..5}; do
    # Generar un numero random
    IMAGEN=$((RANDOM % 3))

    if [ "$IMAGEN" = "0" ]; then
        # Imagen 1
        docker run -d roldyoran/go-client
    else
        if [ "$IMAGEN" = "1" ]; then
            # Imagen 2
            docker run -d alpine sh -c "while true; do echo '2^20' | bc > /dev/null; sleep 2; done"
        else
            # Imagen 3
            docker run -d alpine sleep 240
        fi
    fi
done