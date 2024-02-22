/invoice?currency=*&amount=*&account=* - to add money
/withdraw?currency=*&amount=*&account=*&accountTo=* - retransfer money

sql:
database name : transaction
sql script:
CREATE TABLE IF NOT EXISTS public.transaction
(
    id integer NOT NULL DEFAULT nextval('transaction_id_seq'::regclass),
    customerid integer,
    num_invoice character varying(18) COLLATE pg_catalog."default",
    currency currency,
    amount real,
    action action,
    to_num_invoice character varying(18) COLLATE pg_catalog."default",
    statustrans status NOT NULL,
    CONSTRAINT transaction_pkey PRIMARY KEY (id),
    CONSTRAINT transaction_num_invoice_fkey FOREIGN KEY (num_invoice)
        REFERENCES public.accounts (num_invoice) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT transaction_to_num_invoice_fkey FOREIGN KEY (to_num_invoice)
        REFERENCES public.accounts (num_invoice) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)


RABBIT:
    host: /
    user: transaction