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
 
create stream battle_calculate as
select
battle_attacker_enriched.battle_id as battle_id
, pokemon_attacker
, pokemon_attacker_level
, ((0.5 * pokemon_attacker_hp * (pokemon_attacker_attack / pokemon_defender_defense) * (pokemon_attacker_level / pokemon_defender_level)) * random()) as pokemon_attacker_outcome
, pokemon_defender
, pokemon_defender_level
, ((0.5 * pokemon_defender_hp * (pokemon_defender_attack / pokemon_attacker_defense) * (pokemon_defender_level / pokemon_attacker_level)) * random()) as pokemon_defender_outcome
from battle_attacker_enriched
join battle_defender_enriched
within 15 seconds
on battle_attacker_enriched.battle_id = battle_defender_enriched.battle_id
emit changes;

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

