# Pokemon Battle Sim

> Demo project to mess around

![header](https://i.imgur.com/0YEUV00.png)

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Apache Kafka](https://img.shields.io/badge/Apache%20Kafka-000?style=for-the-badge&logo=apachekafka)

# Project Overview

The point of this project was to use [ksqlDB](https://ksqldb.io) to simulate and aggregate synthetic Pokemon data in real-time - as if we were observing results from a real life scenario. Initially, I uploaded the list of all Pokemon from Gen 1 and their attributes into a table in ksqlDB since the data is static. From there I incrementally built the streaming app using a series of stream transformations - each one their own Kafka topic. The battles were created using a Go script. The final table contains each pokemon and their victories aggregated and windowed into 1 minute intervals.

In some ways ksqlDB is similar to [DBT](https://www.getdbt.com); however, the streaming angle provides lots of quirks. For one, the idea of `key` - mandatory for a table and optional for a stream. Additionally, this `key` must come from the key field within a Kafka topic - not the value. Furthermore, in order to join streams they must have the same `key` value, which is why I had to repartition on `battle_id` within `battle_xxx_enriched` streams. Also, the rules around when and how you can aggregate are interesting: every aggregate produces a table - never a stream, certain aggregate functions only work for a stream or table and aggregating on multiple columns can only be done when they are both primary keys.

The damage logic is also a little loose considering I wanted the battle to end in one turn. The logic is largely borrowed from [Pokemon Go](https://gamerant.com/pokemon-go-damage-calculation-explanation-guide/).

# Architecture Overview

```text
+----------+     +---------+     +----------+
|          | --> |         |     |          |
|  GOLANG  | --> |  KAFKA  | --> |  KSQLDB  |
|          | --> |         |     |          |
+----------+     +---------+     +----------+
```

# Installation

* [Install Kafka](https://tecadmin.net/how-to-install-apache-kafka-on-ubuntu-22-04/)
* [Install KsqlDB](https://ksqldb.io/quickstart-standalone-deb.html#quickstart-content)

# Queries

* Create table for all Pokemon data
```sql
create table pokemon (
select
  pokemon varchar primary key
  , type_one varchar
  , type_two varchar
  , total int
  , hp int
  , attack int
  , defense int
  , sp_attack int
  , sp_def int
  , speed int
  ) 
with (
  kafka_topic = 'pokemon'
  , value_format = 'json'
);
```

* Create initial ingestion streams for each side
```sql
create stream battle_attacker (
    pokemon varchar key,
    pokemon_level int,
    battle_id int
) with (
    kafka_topic = 'battle_attacker',
    value_format = 'json',
    partitions = 1
);

create stream battle_defender (
    pokemon varchar key,
    pokemon_level int,
    battle_id int
) with (
    kafka_topic = 'battle_defender',
    value_format = 'json',
    partitions = 1
);
```

* Enrich raw topics with Pokemon details and repartition on `battle_id` so we can join streams together
```sql
create stream battle_attacker_enriched
with (kafka_topic = 'battle_attacker_enriched') as
select
    pokemon.pokemon as pokemon_attacker
    , cast(battle_attacker.pokemon_level as double) as pokemon_attacker_level
    , pokemon.type_one  as pokemon_attacker_type
    , cast(pokemon.hp as double) as pokemon_attacker_hp
    , cast(pokemon.attack as double) as pokemon_attacker_attack
    , cast(pokemon.defense as double) as pokemon_attacker_defense
    , battle_attacker.battle_id
    from battle_attacker
    join pokemon
    on battle_attacker.pokemon = pokemon.pokemon
    partition by battle_id
emit changes;


create stream battle_defender_enriched
with (kafka_topic = 'battle_defender_enriched') as
select
    pokemon.pokemon as pokemon_defender
    , cast(battle_defender.pokemon_level as double) as pokemon_defender_level
    , pokemon.type_one as pokemon_defender_type
    , cast(pokemon.hp as double) as pokemon_defender_hp
    , cast(pokemon.attack as double) as pokemon_defender_attack
    , cast(pokemon.defense as double) as pokemon_defender_defense
    , battle_defender.battle_id
    from battle_defender
    join pokemon
    on battle_defender.pokemon = pokemon.pokemon
    partition by battle_id
 emit changes;
 ```

* Join streams and calculate damage with window of 15 seconds
```sql
create stream battle_calculate as
select
  battle_attacker_enriched.battle_id as battle_id
  , pokemon_attacker
  , pokemon_attacker_level
  , (
    (0.5 * pokemon_attacker_hp *
    (pokemon_attacker_attack / pokemon_defender_defense) *
    (pokemon_attacker_level / pokemon_defender_level)) *
    random()
  ) as pokemon_attacker_outcome
  , pokemon_defender
  , pokemon_defender_level
  , (
  (0.5 * pokemon_defender_hp *
  (pokemon_defender_attack / pokemon_attacker_defense) *
  (pokemon_defender_level / pokemon_attacker_level)) *
  random()
  ) as pokemon_defender_outcome
  from battle_attacker_enriched
  join battle_defender_enriched
  within 15 seconds
  on battle_attacker_enriched.battle_id = battle_defender_enriched.battle_id
emit changes;
```

* Figure out winner
```sql
create stream battle_outcome as
select
  case
     when pokemon_attacker_outcome > pokemon_defender_outcome then pokemon_attacker
     when pokemon_attacker_outcome < pokemon_defender_outcome then pokemon_defender
     else 'tie'
  end as pokemon_winner
  , pokemon_attacker_level
  , pokemon_defender_level
  , battle_id
  from battle_calculate
emit changes;
```

* Aggregate results for each Pokemon into 1 minute windows -> `windowstart` and `windowend` fields put into table automatically
```sql
create table pokemon_victories as 
select
  pokemon_winner as pokemon
  , count(*) as victories
  , avg(pokemon_attacker_level) as avg_winner_level
  , avg(pokemon_defender_level) as avg_loser_level
  from battle_outcome
  window tumbling (size 1 minute, grace period 30 seconds)
  group by pokemon_winner
emit changes;
```

# Versions

* Kafka -> v 3.2.1
* KsqlDB -> v 0.27.2
* Go -> v 1.18.1
