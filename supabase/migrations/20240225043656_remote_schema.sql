create table "public"."request_errors" (
    "id" bigint generated by default as identity not null,
    "created_at" timestamp with time zone not null default now(),
    "url" text,
    "status" smallint
);


alter table "public"."request_errors" enable row level security;

CREATE UNIQUE INDEX request_errors_pkey ON public.request_errors USING btree (id);

alter table "public"."request_errors" add constraint "request_errors_pkey" PRIMARY KEY using index "request_errors_pkey";

grant delete on table "public"."request_errors" to "anon";

grant insert on table "public"."request_errors" to "anon";

grant references on table "public"."request_errors" to "anon";

grant select on table "public"."request_errors" to "anon";

grant trigger on table "public"."request_errors" to "anon";

grant truncate on table "public"."request_errors" to "anon";

grant update on table "public"."request_errors" to "anon";

grant delete on table "public"."request_errors" to "authenticated";

grant insert on table "public"."request_errors" to "authenticated";

grant references on table "public"."request_errors" to "authenticated";

grant select on table "public"."request_errors" to "authenticated";

grant trigger on table "public"."request_errors" to "authenticated";

grant truncate on table "public"."request_errors" to "authenticated";

grant update on table "public"."request_errors" to "authenticated";

grant delete on table "public"."request_errors" to "service_role";

grant insert on table "public"."request_errors" to "service_role";

grant references on table "public"."request_errors" to "service_role";

grant select on table "public"."request_errors" to "service_role";

grant trigger on table "public"."request_errors" to "service_role";

grant truncate on table "public"."request_errors" to "service_role";

grant update on table "public"."request_errors" to "service_role";

create policy "All insert"
on "public"."request_errors"
as permissive
for insert
to anon
with check (true);



