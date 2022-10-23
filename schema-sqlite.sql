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
    description       varchar(400)     ,
    user_id    varchar(21)    not null,
    rule_id    varchar(22)    not null,
    score      int not null,
    moves      int    not null,
    play_state INT             not null,
    data       blob  not null,
    primary key (id),
    foreign key (rule_id) references rule
);


create table if not exists game_history
(
    created_at datetime        not null,
    game_id    varchar(21)    not null,
    move       int    not null,
    kind       INT             not null,
    points     INT             not null,
    data       blob,
    primary key (move, game_id),
    foreign key (game_id) references game
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

create unique index if not exists game_id_move
    on game_history (game_id, move);

create unique index if not exists active_game_id
    on user (active_game_id);

create unique index if not exists slug
    on rule (slug);

