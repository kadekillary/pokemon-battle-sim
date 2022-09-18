.PHONY: remove_kafka ksql kserver list

ksql:
	/usr/bin/ksql http://0.0.0.0:8088

kserver:
	sudo /usr/bin/ksql-server-start /etc/ksqldb/ksql-server.properties

list:
	kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list

remove_kafka:
	kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic battle_attacker
	kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic battle_attacker_enriched
	kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic battle_defender
	kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic battle_defender_enriched
	kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic BATTLE_CALCULATE
	kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic BATTLE_OUTCOME
	kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic POKEMON_VICTORIES
