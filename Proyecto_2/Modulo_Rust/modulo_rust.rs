use kernel::prelude::*;

module! {
    type: ModuloRust, name: b"rust_module", author: b"Sebastián Gómez - 201602929", description: b"Módulo de Linux en Rust", license: b"GPL"}

struct ModuloRust;

impl kernel::Module for ModuloRust {
    fn init(_module: &'static kernel::ThisModule) -> Result<Self> {
        pr_info!("Hola Mundo 201602929\n");
        Ok(ModuloRust)
    }
}

impl Drop for ModuloRust {
    fn drop(&mut self) {
        pr_info!("Adiós Mundo 201602929\n");
    }
}
