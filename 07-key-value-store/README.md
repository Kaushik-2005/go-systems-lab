# Key-Value Store

## What it is

An in-process store that maps string keys to string values and persists mutations in an append-only JSON log.

## What it solves

It provides simple `Put`, `Get`, and `Delete` operations while keeping data recoverable after the process restarts.

## Project structure

```text
store/store.go        map, append-only records, replay, compaction
cmd/basic/             in-memory CRUD demonstration
cmd/concurrent/        concurrent access demonstration
cmd/persistent/        persistence, recovery, and compaction demonstration
```

## Run it

```powershell
go run .\cmd\basic\main.go
go run .\cmd\concurrent\main.go
go run .\cmd\persistent\main.go
```

The persistent store writes `PUT` and `DELETE` records, replays them during `Open`, and compacts the log to retain only the latest live values.
