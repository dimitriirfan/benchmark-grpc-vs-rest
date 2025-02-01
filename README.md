# gRPC vs REST with JSON Performance Benchmark

This repository contains a performance benchmark comparison between gRPC and REST API implementations.

## Description

This project provides a comprehensive benchmark suite to compare performance between gRPC and REST with JSON APIs. It includes test fixtures and utilities to measure latency between client and server communications using both protocols.

The benchmark focuses on:
- Client-to-server request latency
- Different payload sizes using a sample population dataset
- Real-world usage scenarios with structured data
- Direct comparison between REST (JSON/HTTP) and gRPC (Protocol Buffers/HTTP2)

## Results

## Project Structure

- `proto/`: Contains Protocol Buffer definitions
- `testutil/fixtures/`: Test data and fixture generation utilities
  - `generate_fixtures.py`: Script to generate test data in both JSON and Protocol Buffer formats
  - `fixtures_population_100.json`: Sample population data in JSON format
  - `fixtures_population_100.pb`: Sample population data in Protocol Buffer format

## General Testing

Tests can be run to compare the performance characteristics of both API implementations.

