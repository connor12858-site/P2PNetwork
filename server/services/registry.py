from repositories import registry_repo


def get_nodes():
    return registry_repo.all_nodes()


def register(node):
    registry_repo.register(node)


def unregister(peer):
    registry_repo.unregister(peer)