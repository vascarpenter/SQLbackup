# SQLbackup

- Oracle SQL backup コマンドラインツール (goで作成)

- ビルド方法 
  - `go build -o SQLbackup`

- 使用方法
  - `./SQLbackup user/pass@tnsaddr [-d] [--tables <name1,name2>] [--tablespace <name>]`
    - `-d` "drop table" を  create tableの前に挿入
    - `--tablespace <name>` table space名を変更 (default: DATA) (このオプションは現在意味がない)
    - `--tables <table1,table2,...>` 出力するテーブル名を変更。","で区切ること。指定しない場合すべてのテーブルを出力。

- 例 `./SQLbackup $OCISTRING -d --tables BLOODTEMP >test.sql`

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

- 変更
  - 20201129 VARCHAR2でも single quote
