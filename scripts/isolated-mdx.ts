import { compile } from "@mdx-js/mdx";
import { getQuickJS, shouldInterruptAfterDeadline } from "quickjs-emscripten";
import remarkGfm from "remark-gfm";

// Only serialized element data leaves this WebAssembly interpreter. No Node,
// filesystem, network, environment or host callbacks are exposed to authored JS.
export async function isolatedMdx(
  mode: string,
  source: string,
  names: string[],
): Promise<unknown> {
  const aliases: {
    local: string;
    imported: string;
    namespace: boolean;
    react: boolean;
  }[] = [];
  function emailImports() {
    return (tree: any) => {
      for (const node of tree.children) {
        if (node.type !== "mdxjsEsm") continue;
        const keep = [];
        for (const statement of node.data.estree.body) {
          if (statement.type !== "ImportDeclaration") {
            keep.push(statement);
            continue;
          }
          const module = statement.source.value;
          if (module !== "@react-email/components" && module !== "react")
            throw new Error("Only React Email and React imports are supported");
          for (const spec of statement.specifiers) {
            const namespace =
              spec.type === "ImportNamespaceSpecifier" ||
              spec.type === "ImportDefaultSpecifier";
            const imported = spec.imported?.name ?? spec.imported?.value ?? "";
            if (
              module === "@react-email/components" &&
              !namespace &&
              !names.includes(imported)
            )
              throw new Error("Unknown React Email component");
            if (
              module === "react" &&
              !namespace &&
              !["createElement", "Fragment"].includes(imported)
            )
              throw new Error("Only static React rendering is supported");
            aliases.push({
              local: spec.local.name,
              imported,
              namespace,
              react: module === "react",
            });
          }
        }
        node.data.estree.body = keep;
      }
    };
  }
  const code = String(
    await compile(source, {
      outputFormat: "function-body",
      format: mode === "markdown" ? "md" : "mdx",
      remarkPlugins: [remarkGfm, emailImports],
    }),
  );
  const globals = aliases
    .map(
      (a) =>
        `const ${a.local}=${a.react ? "__React" : "__components"}${a.namespace ? "" : `[${JSON.stringify(a.imported)}]`};${!a.react ? `__components[${JSON.stringify(a.local)}]=${a.local};` : ""}`,
    )
    .join("\n");
  const program = `(()=>{
 const __fragment='__email_fragment';
 const __jsx=(type,props)=>{if(type===__fragment)return props?.children??null;if(typeof type==='function')return type(props||{});if(typeof type!=='string')throw Error('Invalid element');return {type,props:props||{}}};
 const __React={Fragment:__fragment,createElement:(type,props,...children)=>__jsx(type,{...props,...(children.length?{children:children.length===1?children[0]:children}:{})})};
 const __components=Object.create(null);
 for(const name of ${JSON.stringify(names)})__components[name]=props=>({type:name,props:props||{}});
 ${globals}
 const __module=(function(){${code}})({jsx:__jsx,jsxs:__jsx,Fragment:__fragment});
 const output=__jsx(__module.default,{components:__components});
 const json=JSON.stringify(output,(_key,value)=>{if(typeof value==='function')throw Error('Function-valued props are not supported');return value});
 if(json.length>1000000)throw Error('Output too large');return json;
 })()`;
  const vm = await getQuickJS();
  const result = vm.evalCode(program, {
    memoryLimitBytes: 32 * 1024 * 1024,
    maxStackSizeBytes: 1024 * 1024,
    shouldInterrupt: shouldInterruptAfterDeadline(Date.now() + 750),
  });
  if (typeof result !== "string") throw new Error("Invalid render tree");
  return JSON.parse(result);
}
