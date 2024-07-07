create table dotabuff_match
(
    id                     bigint,
    dire_heroes            text,
    radiant_heroes         text,
    radiant_won            boolean,
    radiant_won_prediction boolean,
    tournament_link        text,
    dire_team_link         text,
    radiant_team_link      text,
    radiant_win_prediction float,
    dire_win_prediction    float,
    algorithm_version      varchar(255),
    PRIMARY KEY (id, algorithm_version)
);