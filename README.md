# SQLbackup

- Oracle SQL backup command line tool (written in go)

- BUILD 
  - `go build -o SQLbackup`

- Usage
  - `./SQLbackup user/pass@tnsaddr [-d] [--tables <name1,name2>] [--tablespace <name>]`
    - `-d` Insert "drop table" before create
    - `--tablespace <name>` Change default table space name (default: DATA, that is used in Autonomous Data Warehouse) (This option is meaningless.)
    - `--tables <table1,table2,...>` Specify table name(s) to export. Separate names by ",".

- eg. `./SQLbackup $OCISTRING -d --tables BLOODTEMP >test.sql`

```sql
DROP TABLE "BLOODTEMP";
CREATE TABLE "ADMIN"."BLOODTEMP" 
   (    "DATE" DATE NOT NULL ENABLE, 
        "TEMP" NUMBER NOT NULL ENABLE
   )  DEFAULT COLLATION "USING_NLS_COMP" SEGMENT CREATION IMMEDIATE 
  PCTFREE 10 PCTUSED 40 INITRANS 10 MAXTRANS 255 
 NOCOMPRESS LOGGING
  STORAGE(INITIAL 65536 NEXT 1048576 MINEXTENTS 1 MAXEXTENTS 2147483645
  PCTINCREASE 0 FREELISTS 1 FREELIST GROUPS 1
  BUFFER_POOL DEFAULT FLASH_CACHE DEFAULT CELL_FLASH_CACHE DEFAULT)
  TABLESPACE "DATA";

SET DEFINE OFF;
Insert Into BLOODTEMP ("DATE","TEMP") VALUES (TO_DATE('20-07-28','RR-MM-DD'),36.5);
...
Insert Into BLOODTEMP ("DATE","TEMP") VALUES (TO_DATE('20-08-14','RR-MM-DD'),36.7);
```

- Changes
  - 20201129 VARCHAR2 data, single quote (was: NVARCHAR)
  - 20201204 changed oracle driver to https://github.com/godror/godror