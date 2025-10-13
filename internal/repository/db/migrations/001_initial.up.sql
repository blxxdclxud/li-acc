CREATE TABLE files (
    Id SERIAL PRIMARY KEY,
    FileName VARCHAR(256),
    File bytea,
    ModifiedDate timestamp
);

CREATE TABLE settings (
    Id SERIAL PRIMARY KEY,
    Emails VARCHAR(65535),
    SenderEmail VARCHAR(65535)
);
