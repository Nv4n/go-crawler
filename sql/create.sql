CREATE TABLE IF NOT EXISTS image_metadata(
    id int generated always as identity ,
    filename varchar(255) unique ,
    title text,
    alt_text varchar(255),
    resolution varchar(20),
    format varchar(10),
 CONSTRAINT PK_IMAGE_METADATA PRIMARY KEY (id)
);