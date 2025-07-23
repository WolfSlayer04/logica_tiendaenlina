package rutas

// RouteVarsKey se usa para acceder a las variables de ruta desde el contexto
var RouteVarsKey = &contextKey{"route_vars"}

type contextKey struct {
    name string
}

func (k *contextKey) String() string {
    return "onlinestore context key: " + k.name
}