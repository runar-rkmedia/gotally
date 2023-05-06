-- This file is generated
-- Please do not edit.
-- The file to edit should be ../schema-sqlite.sql
create table if not exists rule
(
    id                varchar(21)     not null,
    slug              varchar(64)      not null,
    created_at        datetime         not null,
    updated_at        datetime,
    mode              INT              not null,
    description       varchar(400)     ,
    size_x            int not null,
    size_y            int not null,
    max_moves            int,
    target_cell_value            int,
    target_score            int,
    recreate_on_swipe BOOLEAN          not null,
    no_reswipe        BOOLEAN          not null,
    no_multiply       BOOLEAN          not null,
    no_addition       BOOLEAN          not null,
    primary key (id),
    constraint my_uniq_id
        unique (slug, description)
);

create table if not exists game
(
    id         varchar(21)    not null,
    created_at datetime default CURRENT_TIMESTAMP not null,
    updated_at datetime,
    name       varchar(80)     ,
    description       varchar(400)     ,
    user_id    varchar(21)    not null,
    rule_id    varchar(21)    not null,
    based_on_game varchar(21),
    template_id    varchar(21),
    score      int not null,
    moves      int    not null,
    play_state INT             not null,
    data       blob  not null,
    data_at_start       blob  not null,
    history       blob,
    primary key (id),
    foreign key (rule_id) references rule,
    foreign key (template_id) references game_template,
    foreign key (user_id) references user
);


create table if not exists user
(
    id             varchar(21) not null,
    created_at     datetime     not null,
    updated_at     datetime,
    username       varchar(21) not null,
    active_game_id varchar(21) not null,
    primary key (id),
    foreign key (active_game_id) references game
);


create table if not exists session
(
    id            varchar(21) not null,
    created_at    datetime     not null,
    updated_at    datetime,
    invalid_after datetime     not null,
    user_id       varchar(21) not null,
    primary key (id),
    foreign key (user_id) references user
);
create table if not exists game_template
(
    id            varchar(21) not null,
    created_at    datetime     not null,
    updated_at    datetime,
    rule_id    varchar(21)    not null,
    created_by    varchar(21)    not null,
    updated_by    varchar(21),
    name       varchar(80)     not null,
    description       varchar(400)     ,
    challenge_number INT,
    ideal_moves INT,
    ideal_score INT,
    data       blob  not null,
    UNIQUE(challenge_number),
    primary key (id),
    foreign key (rule_id) references rule,
    foreign key (created_by) references user,
    foreign key (updated_by) references user
);


create unique index if not exists active_game_id
    on user (active_game_id);

create unique index if not exists slug
    on rule (slug);
create unique index if not exists game_template_challenge_number
    on game_template (challenge_number);

