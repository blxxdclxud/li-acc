CREATE TABLE files (
    Id INT PRIMARY KEY,
    FileName VARCHAR(256),
    File BLOB,
    ModifiedDate DATETIME
);

CREATE TABLE settings (
    Id INT PRIMARY KEY,
    ReceiptFile BLOB,
    Emails BLOB,
    QrPattern VARCHAR(65535),
    SenderEmail VARCHAR(65535)
);
