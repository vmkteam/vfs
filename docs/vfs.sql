Create table "vfsFiles"
(
    "fileId" Serial NOT NULL,
    "folderId" Integer NOT NULL,
    "title" Varchar(255) NOT NULL,
    "path" Varchar(255) NOT NULL,
    "params" Text,
    "isFavorite" Boolean Default false,
    "mimeType" Varchar(255) NOT NULL,
    "fileSize" Integer Default 0,
    "fileExists" Boolean NOT NULL Default true,
    "statusId" Integer NOT NULL,
    "createdAt" Timestamp NOT NULL Default now(),
    primary key ("fileId")
) Without Oids;

Create table "vfsFolders"
(
    "folderId" Serial NOT NULL,
    "parentFolderId" Integer,
    "title" Varchar(255) NOT NULL,
    "isFavorite" Boolean Default false,
    "createdAt" Timestamp NOT NULL Default now(),
    "statusId" Integer NOT NULL,
    primary key ("folderId")
) Without Oids;

Create table "vfsHashes"
(
    "hash" varchar(40) not null,
    "namespace" varchar(32) not null,
    "extension" varchar(4) not null,
    "fileSize" Integer not null Default 0,
    "width" int not null default 0,
    "height" int not null default 0,
    "blurhash" text,
    "createdAt" Timestamp with time zone NOT NULL Default now(),
    "indexedAt" Timestamp with time zone,
    primary key ("hash","namespace")
) Without Oids;


Create index "IX_FK_vfsFoldersFolderId_vfsFolders" on "vfsFolders" ("parentFolderId");
Alter table "vfsFolders" add  foreign key ("parentFolderId") references "vfsFolders" ("folderId") on update restrict on delete restrict;
Create index "IX_FK_vfsFilesFolderId_vfsFiles" on "vfsFiles" ("folderId");
Alter table "vfsFiles" add  foreign key ("folderId") references "vfsFolders" ("folderId") on update restrict on delete restrict;
Create index "IX_vfsHashes_indexedAt" on "vfsHashes" ("indexedAt");
