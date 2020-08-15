# SQLbackup

- SQL backup command line tool (written in go)

- BUILD 
  - `go build .`
  - `mv m sqlbackup`

- Usage
  - `sqlbackup user/pass@tnsaddr [-d]`
    - `-d` drop table before create
    - `--tablespace <name>` change default table space name (default: DATA, that is used in Autonomous Data Warehouse)
  - write SQL dump for all table in database user

- caution
  - indexes are ignored
  
- eg. `sqlbackup $OCISTRING -d >test.sql`

```sql
DROP TABLE "DATETEST";

  CREATE TABLE "ADMIN"."DATETEST" 
   (	"DATE1" DATE, 
	"DATEBOOL" NUMBER
   )  DEFAULT COLLATION "USING_NLS_COMP" SEGMENT CREATION IMMEDIATE 
  PCTFREE 10 PCTUSED 40 INITRANS 1 MAXTRANS 255 
 NOCOMPRESS LOGGING
  STORAGE(INITIAL 65536 NEXT 1048576 MINEXTENTS 1 MAXEXTENTS 2147483645
  PCTINCREASE 0 FREELISTS 1 FREELIST GROUPS 1
  BUFFER_POOL DEFAULT FLASH_CACHE DEFAULT CELL_FLASH_CACHE DEFAULT)
  TABLESPACE "DATA";

SET DEFINE OFF;
Insert Into DATETEST ("DATE1","DATEBOOL") VALUES (TO_DATE('20-06-27','RR-MM-DD'),1);
Insert Into DATETEST ("DATE1","DATEBOOL") VALUES (TO_DATE('20-06-26','RR-MM-DD'),2);

```
