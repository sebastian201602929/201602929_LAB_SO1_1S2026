#include <linux/init.h> // Contiene las macros __init y __exit
#include <linux/module.h> // Creación de módulos
#include <linux/kernel.h> // Impresión de mensajes en el kernel
#include <linux/proc_fs.h> // Crear archivos en /proc
#include <linux/seq_file.h> // Escribir en archivos
#include <linux/mm.h> // Manejo de Memoria
#include <linux/sched.h>  // Estructura task_struct
#include <linux/timer.h> // Timers
#include <linux/jiffies.h> // Jiffies, ticks del sistema

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Sebastián Gómez Lavarreda - 201602929");
MODULE_DESCRIPTION("Módulo que captura métricas del sistema. Parte del Proyecto 2 de Laboratorio de Sistemas Operativos 1 en el 1 semestre de 2026.");
MODULE_VERSION("1.0");

#define PROC_NAME "continfo_pr2_so1_201602929" // Nombre del archivo en /proc

// Función para obtener la línea de comandos de un proceso
static char *get_process_cmdline(struct task_struct *task) {
    struct mm_struct *mm;
    char *cmdline, *p;
    unsigned long arg_start, arg_end, env_start;
    int i, len;

    // Reservamos memoria para la línea de comandos
    cmdline = kmalloc(256, GFP_KERNEL);
    if (!cmdline) return NULL;

    // Obtenemos la información de memoria
    mm = get_task_mm(task);
    if (!mm) {
        kfree(cmdline);
        return NULL;
    }

    // Bloqueo de lectura
    down_read(&mm->mmap_lock);

    // Direccion de inicio y fin de los argumentos y el entorno
    arg_start = mm->arg_start;
    arg_end = mm->arg_end;
    env_start = mm->env_start;

    // Libera bloqueo de lectura
    up_read(&mm->mmap_lock);

    // Longitud de la línea de comandos
    len = arg_end - arg_start;

    if (len > 255)
        len = 255;

    // Línea de comandos de la memoria virtual del proceso
    if (access_process_vm(task, arg_start, cmdline, len, 0) != len) {
        mmput(mm);
        kfree(cmdline);
        return NULL;
    }

    // Caracter nulo al final
    cmdline[len] = '\0';

    // Reemplaza caracteres nulos por espacios
    p = cmdline;
    for (i = 0; i < len; i++)
        if (p[i] == '\0')
            p[i] = ' ';

    // Libera la estructura
    mmput(mm);

    return cmdline;
}

// Funcion para obtener informacion del sistema y escribir en /proc
static int sysinfo_show(struct seq_file *m, void *v) {
    struct sysinfo si;
    struct task_struct *task;
    unsigned long t_jiffies = jiffies;
    char *cmdline = NULL;

    // Obtener información de la memoria
    si_meminfo(&si);

    // Imprimir en el archivo en /proc
    seq_printf(m, "Total de memoria RAM: %lu MB\n", (si.totalram << (PAGE_SHIFT - 10)) / 1024);
    seq_printf(m, "Memoria RAM libre: %lu MB\n", (si.freeram << (PAGE_SHIFT - 10)) / 1024);
    seq_printf(m, "Memoria RAM en uso: %lu MB\n", ((si.totalram - si.freeram) << (PAGE_SHIFT - 10)) / 1024);

    seq_printf(m, "\n");

    // Procesos relacionados a contenedores
    seq_printf(m, "PID\tNombre\tCMD\tVSZ\tRSS\t%% RAM\t%% CPU\t\n");
    for_each_process(task) {
        if (strcmp(task->comm, "containerd-shim") == 0) {
            unsigned long vsz = 0;
            unsigned long rss = 0;
            unsigned long mem_usage = 0;
            unsigned long cpu_usage = 0;
            char *cmdline = NULL;

            // Obtiene la linea de comando que se ejecuto
            cmdline = get_process_cmdline(task);

            // Obtiene y calcula uso de memoria y CPU
            if (task->mm) {
                vsz = task->mm->total_vm << (PAGE_SHIFT - 10);
                rss = get_mm_rss(task->mm) << (PAGE_SHIFT - 10);

                mem_usage = (rss * 100) / (si.totalram << (PAGE_SHIFT - 10));
                cpu_usage = (task->utime + task->stime) * 10000 / t_jiffies;
            }

            struct task_struct *child;
            list_for_each_entry(child, &task->children, sibling) {
                if (child->mm) {
                    // Cálculo del uso de memoria
                    rss = get_mm_rss(child->mm) << (PAGE_SHIFT - 10);
                    mem_usage = (rss * 100) / (si.totalram << (PAGE_SHIFT - 10));

                    // Cálculo del uso de CPU en función de jiffies
                    unsigned long total_time_child = child->utime + child->stime;
                    u64 time_since_start_jiffies = get_jiffies_64();

                    if (time_since_start_jiffies > 0 && t_jiffies > 0) {
                        // Calcula el uso de CPU como un porcentaje en base a jiffies
                        cpu_usage = (total_time_child * 100) / t_jiffies;
                        cpu_usage = cpu_usage / num_online_cpus();
                    } else {
                        cpu_usage = 44;
                    }

                    break;
                }
            }

            seq_printf(m, "%d\t%s\t%s\t%lu\t%lu\t%lu.%lu\t%lu.%lu\n", task->pid, task->comm, cmdline ? cmdline : "N/A", vsz, rss, mem_usage /10, mem_usage %10, cpu_usage / 100, cpu_usage % 100);

            if (cmdline) kfree(cmdline);
        }
    }

    if (cmdline) kfree(cmdline);

    return 0;
}

static int sysinfo_open(struct inode *inode, struct file *file) {
    return single_open(file, sysinfo_show, NULL);
}

static const struct proc_ops sysinfo_ops = {
    .proc_open = sysinfo_open,
    .proc_read = seq_read,
};

static int __init modulo_init(void) {
    proc_create(PROC_NAME, 0, NULL, &sysinfo_ops);
    printk(KERN_INFO "Módulo cargado: Captura de métricas del sistema.\n");
    return 0;
}

static void __exit modulo_exit(void) {
    remove_proc_entry(PROC_NAME, NULL);
    printk(KERN_INFO "Módulo descargado: Captura de métricas del sistema.\n");
}

module_init(modulo_init);
module_exit(modulo_exit);