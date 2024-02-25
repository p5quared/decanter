alter table "public"."requests" alter column "size" set data type bigint using "size"::bigint;

create policy "Disable deletes"
on "public"."requests"
as permissive
for delete
to anon
using (false);


create policy "Everyone can insert"
on "public"."requests"
as permissive
for insert
to anon
with check (true);


create policy "No selects"
on "public"."requests"
as permissive
for select
to anon
using (false);


create policy "No updates"
on "public"."requests"
as permissive
for update
to anon
using (false)
with check (false);




