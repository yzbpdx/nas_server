create user if not exists 'nas'@'localhost' identified by 'server';
GRANT ALL PRIVILEGES ON *.* TO 'nas'@'localhost' with grant option;
flush privileges;

create database if not exists nas_server;

use nas_server;

create table if not exists user (
    id int primary key auto_increment not null,
    username varchar(255) not null unique,
    password varchar(255) not null
);