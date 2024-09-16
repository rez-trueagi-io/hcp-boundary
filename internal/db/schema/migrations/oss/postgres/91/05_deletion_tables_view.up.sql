-- Copyright (c) HashiCorp, Inc.
-- SPDX-License-Identifier: BUSL-1.1

begin;
  -- Originially added in 81/01_deleted_tables_and_triggers.up.sql
  -- This is being replaced with a view.
  drop function get_deletion_tables;

  create view deletion_table as
    select c.relname as tablename
      from pg_catalog.pg_class c
     where c.relkind in ('r')
       and c.relname operator(pg_catalog.~) '^(.+_deleted)$' collate pg_catalog.default
       and pg_catalog.pg_table_is_visible(c.oid);
commit;
