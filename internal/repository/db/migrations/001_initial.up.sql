CREATE TABLE files (
    Id SERIAL PRIMARY KEY,
    FileName VARCHAR(256),
    File bytea,
    ModifiedDate timestamp
);

CREATE TABLE settings (
    Id SERIAL PRIMARY KEY,
    ReceiptFile bytea,
    Emails bytea,
    QrPattern VARCHAR(65535),
    SenderEmail VARCHAR(65535)
);
