#include "czmq.h"

void startExternalServer()
{
    zsock_t *server = zsock_new (ZMQ_SERVER);
    assert (server);
    zsock_bind (server, "tcp://127.0.0.1:31337");
    char *msg = zstr_recv (server);
	  zstr_send (server, "WORLD");
    zsock_destroy (&server);
}

void startExternalRouter()
{
    zsock_t *router = zsock_new (ZMQ_ROUTER);
    assert (router);
    int rc = zsock_bind (router, "tcp://127.0.0.1:%d", 31340);
    assert (rc == 31340);
    assert (streq (zsock_endpoint (router), "tcp://127.0.0.1:31340"));

    char *ident = zstr_recv (router);

    zstr_sendm (router, ident);
    zstr_send (router, "WORLD");

    zsock_destroy (&router);
}
