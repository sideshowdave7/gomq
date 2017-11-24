#include "czmq.h"

void startExternalServer()
{
    zsock_t *server = zsock_new (ZMQ_SERVER);
    assert (server);
    int rc = zsock_bind (server, "tcp://127.0.0.1:31337");
    assert (rc == 31337);
    assert (streq (zsock_endpoint (server), "tcp://127.0.0.1:31337"));

    char *msg = zstr_recv (server);
    assert(streq (msg, "HELLO"));

    printf("hi");
	  zstr_send(server, "WORLD");
    zsock_destroy (&server);
}

void startExternalRouter()
{
    zsock_t *router = zsock_new (ZMQ_ROUTER);
    assert (router);
    int rc = zsock_bind (router, "tcp://127.0.0.1:%d", 31340);
    assert (rc == 31340);
    assert (streq (zsock_endpoint (router), "tcp://127.0.0.1:31340"));

    char *ident;
    zstr_recvx(router, &ident, NULL);

    // Ensure unregistered messages are not received
    zstr_sendx(router, "unregistered_dealer_id", "TEST", NULL);

    zstr_sendm(router, ident);
    zstr_send(router, "WORLD");

    char *id, *first, *second;
    zstr_recvx(router, &id, &first, &second, NULL);

    zstr_sendx(router, ident, "WORLD", "HELLO", NULL);

    zsock_destroy (&router);
}
