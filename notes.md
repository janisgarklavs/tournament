# GET /take
playerId string
points float

# GET /fund
playerId string
points float

# GET /announceTournament
tournamentId string
deposit float

# GET /joinTournament
tournamentId string
playerId string
backerId string (allow multiples)

# POST /resultTournament
```json
{
    "tournamentId":"1", "winners": [
        {"playerId": "P1", "prize": 500},
        ...
    ]
}
```
# GET /balance
playerId string

```json
{"playerId": "P1", "balance": 450.00}
```

# GET /reset
resets db



#game scenario
-there are some players you can either fund or take money from them
-you can announce new tournament with id and required entry fee
-players can join tournament either by themselves depositing entry fee
-players can join tournament backed by other users spliting entry fee in even parts
-user balance allways above zero
-no points are lost due to unexpected errors (transactions)

-postgres tables
-all balance and deposits are 2 decimal place floats converted to ints ( balance * 100)

player (id string, balance int) (enforce constraint on balance for positive values)
tournament (id string unique PK, deposit int, finished bool) (dont accept more joins after finished)
tournament_entries (serial, tournament_id, user_id, backing_id) (user_id cannot be equal backer_id)