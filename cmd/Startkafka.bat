@echo off
chcp 1251

echo Запуск ZooKeeper в новом окне...
start cmd /c "C:\kafka\bin\windows\zookeeper-server-start.bat C:\kafka\config\zookeeper.properties"

echo Запуск Kafka broker в новом окне...
start cmd /c "C:\kafka\bin\windows\kafka-server-start.bat C:\kafka\config\server.properties"

echo Запуск AKHQ в новом окне...
start cmd /c "java -Dmicronaut.config.files=C:\akhq\application.yml -jar C:\akhq\akhq-0.25.1-all.jar" || echo "Ошибка запуска AKHQ. Проверьте конфигурацию."

echo Все сервисы Kafka и AKHQ запущены.
pause > nul