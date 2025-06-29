#!/bin/bash

cd sql/schema
echo "Initiating up migration..."
goose postgres "postgres://postgres:postgres@localhost:5432/gator" up
echo "Up migration complete"
cd ../..