(function () {
    const canvas = document.getElementById('shader-bg');
    if (!canvas) { console.error('[shader] canvas#shader-bg not found'); return; }

    const isMobile = /Mobi|Android/i.test(navigator.userAgent);
    if (isMobile) {
        canvas.style.display = 'none';
        document.body.classList.add('shader-active');
        return;
    }

    const gl = canvas.getContext('webgl2');
    if (!gl) { console.error('[shader] WebGL2 not available'); return; }
    console.log('[shader] WebGL2 context obtained');

    function resize() {
        canvas.width  = window.innerWidth;
        canvas.height = window.innerHeight;
        gl.viewport(0, 0, canvas.width, canvas.height);
    }
    window.addEventListener('resize', resize);
    resize();

    const vertSrc =
`#version 300 es
in vec2 aPos;
void main() { gl_Position = vec4(aPos, 0.0, 1.0); }`;

    const fragSrc =
`#version 300 es
precision highp float;
precision highp int;

uniform float uTime;
uniform vec2  uResolution;
uniform float uRandom;

out vec4 fragColor;

const uint antiA = 2u;

const vec3 mocha_base     = vec3(30.0,30.0,46.0)/255.0;
const vec3 mocha_mantle   = vec3(24.0,24.0,37.0)/255.0;
const vec3 mocha_crust    = vec3(17.0,17.0,27.0)/255.0;
const vec3 mocha_surface0 = vec3(49.0,50.0,68.0)/255.0;
const vec3 mocha_surface1 = vec3(69.0,71.0,90.0)/255.0;
const vec3 mocha_surface2 = vec3(88.0,91.0,112.0)/255.0;
const vec3 mocha_overlay0 = vec3(108.0,112.0,134.0)/255.0;
const vec3 mocha_overlay1 = vec3(127.0,132.0,156.0)/255.0;
const vec3 mocha_overlay2 = vec3(147.0,153.0,178.0)/255.0;
const vec3 mocha_subtext0 = vec3(166.0,173.0,200.0)/255.0;

const vec3 bgColors[4] = vec3[4](mocha_crust,mocha_mantle,mocha_base,mocha_surface0);
const vec3 sfColors[14] = vec3[14](
    mocha_surface0,mocha_surface1,mocha_surface2,mocha_overlay0,
    mocha_overlay1,mocha_overlay2,mocha_subtext0,mocha_surface1,
    mocha_surface2,mocha_overlay0,mocha_surface0,mocha_overlay1,
    mocha_surface1,mocha_surface2);

const float ERR = 1e10;
mat3  fA,cA,fA2,cA2;
vec3  fB,cB,fB2,cB2;
float fC,cC,fC2,cC2;
vec3  bgCol,bgCol2,sfCol,sfCol2;

vec3 hash3(uint n){
    n=(n<<13U)^n;
    n=n*(n*n*15731U+789221U)+1376312589U;
    uvec3 k=n*uvec3(n,n*16807U,n*48271U);
    return vec3(k&uvec3(0x7fffffffU))/float(0x7fffffff);
}

mat2x3 boxM(uint n){
    vec3 U=hash3(n),V=hash3(n+2568758767u);
    U=sqrt(-2.0*log(U));
    V*=6.28318530;
    return mat2x3(U*cos(V),U*sin(V));
}

void setParams(uint n,out mat3 oFA,out mat3 oCA,out vec3 oFB,out vec3 oCB,
               out float oFC,out float oCC,out vec3 oBg,out vec3 oSf){
    n*=100u;
    for(uint i=0u;i<3u;++i){
        mat2x3 t=boxM(n++);
        oFA[i]=t[0]; oCA[i]=t[1];
        oFA[i][i]/=sqrt(2.0); oCA[i][i]/=sqrt(2.0);
    }
    oFB=2.0*(hash3(n++)-1.0);
    oCB=2.0*(hash3(n++)-1.0);
    oFA*=0.3; oCB*=0.3; oCA*=0.2;
    oCC=0.0; oFC=0.0;
    oBg=bgColors[n%4u];
    oSf=sfColors[(n+7u)%14u];
}

void set(uint n,float blend){
    setParams(n,  fA, cA, fB, cB, fC, cC, bgCol,  sfCol);
    setParams(n+1u,fA2,cA2,fB2,cB2,fC2,cC2,bgCol2,sfCol2);
    float ts=0.05;
    float t=0.0;
    if(blend>ts){
        t=(blend-ts)/(1.0-ts);
        t=t*t*t*(t*(t*6.0-15.0)+10.0);
    }
    for(int i=0;i<3;i++){fA[i]=mix(fA[i],fA2[i],t);cA[i]=mix(cA[i],cA2[i],t);}
    fB=mix(fB,fB2,t); cB=mix(cB,cB2,t);
    fC=mix(fC,fC2,t); cC=mix(cC,cC2,t);
    bgCol=mix(bgCol,bgCol2,t); sfCol=mix(sfCol,sfCol2,t);
}

float evalQ(vec3 x,mat3 A,vec3 B,float C){return dot(x,A*x)+dot(B,x)+C;}
vec3  gradQ(vec3 x,mat3 A,vec3 B){return B+A*x+x*A;}
vec3  paramQ(vec3 x,vec3 d,mat3 A,vec3 B,float C){
    return vec3(evalQ(x,A,B,C),dot(gradQ(x,A,B),d),dot(d,A*d));
}

vec2 solve(vec3 p){
    float a=p.z,b=p.y,c=p.x;
    float disc=b*b-4.0*a*c;
    if(disc<0.0) return vec2(ERR);
    float sq=sqrt(disc);
    vec2 tmp;
    tmp.x=(-b-sign(b)*sq)/(2.0*a);
    if(abs(tmp.x)<1e-10) return vec2(ERR);
    tmp.y=c/(a*tmp.x);
    if(tmp.y<tmp.x) tmp=tmp.yx;
    if(tmp.x<0.0)   tmp=vec2(tmp.y,ERR);
    if(tmp.x<0.0)   tmp=vec2(ERR);
    return tmp;
}

void main(){
    const float duration=50.0,speed=0.2,delta=1.5,focal=1.3;
    float tc=mod(uTime,duration);
    uint  seed=uint(uTime/duration)+uint(uRandom*1000.0);
    set(seed,tc/duration);
    fC-=uTime*speed;

    vec2 uv=(2.0*gl_FragCoord.xy-uResolution)/uResolution.y;
    float pixel=1.0/uResolution.y;
    vec3 ro=vec3(0.0,0.0,4.0);
    vec3 tot=vec3(0.0);

    for(uint nAA=0u;nAA<antiA*antiA;++nAA){
        float step=1.0/float(antiA);
        vec2 uv1=vec2(float(nAA/antiA),float(nAA%antiA))*step+0.5*step;
        uv1=(uv1-0.5)*2.0*pixel;
        vec3 rd=normalize(vec3(uv+uv1,-focal));

        vec3 fPar=paramQ(ro,rd,fA,fB,fC);
        vec3 cPar=paramQ(ro,rd,cA,cB,cC);
        vec2 cI=solve(cPar);
        vec3 col;

        if(cI.x<ERR){
            float hit=evalQ(ro+rd*cI.x,fA,fB,fC);
            vec2 pot=delta*(floor(hit/delta)+vec2(0.0,1.0));
            vec4 fI=vec4(solve(fPar-vec3(pot.x,0.0,0.0)),
                         solve(fPar-vec3(pot.y,0.0,0.0)));
            float t=ERR;
            for(int i=0;i<4;++i)
                if(fI[i]<min(t,cI.y)&&fI[i]>cI.x) t=fI[i];
            vec3 pos=ro+t*rd;
            vec3 fg=gradQ(pos,fA,fB)/delta;
            vec3 cg=gradQ(pos,cA,cB)/abs(evalQ(pos,cA,cB,cC));
            float occ=length(fg)/length(cg);
            occ=sqrt(occ);
            occ=1.0-occ/sqrt(1.0+occ*occ);
            occ=sqrt(occ);
            col=t<ERR?sfCol*occ:bgCol;
            col=mix(col,sfCol,smoothstep(-2.0*pixel,-pixel,-1.0/cI.x));
        } else {
            col=bgCol;
        }
        tot+=col;
    }
    tot/=float(antiA*antiA);
    tot=pow(clamp(tot,0.0,1.0),vec3(1.0/2.2));
    fragColor=vec4(tot,1.0);
}`;

    function compile(type, src, name) {
        const s = gl.createShader(type);
        gl.shaderSource(s, src);
        gl.compileShader(s);
        if (!gl.getShaderParameter(s, gl.COMPILE_STATUS)) {
            console.error('[shader] ' + name + ' compile error:\n' + gl.getShaderInfoLog(s));
            gl.deleteShader(s);
            return null;
        }
        console.log('[shader] ' + name + ' compiled OK');
        return s;
    }

    const vs = compile(gl.VERTEX_SHADER,   vertSrc, 'vertex');
    const fs = compile(gl.FRAGMENT_SHADER, fragSrc, 'fragment');
    if (!vs || !fs) return;

    const prog = gl.createProgram();
    gl.attachShader(prog, vs);
    gl.attachShader(prog, fs);
    gl.linkProgram(prog);
    if (!gl.getProgramParameter(prog, gl.LINK_STATUS)) {
        console.error('[shader] link error:\n' + gl.getProgramInfoLog(prog));
        return;
    }
    console.log('[shader] program linked OK');

    const vao = gl.createVertexArray();
    gl.bindVertexArray(vao);
    const buf = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buf);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1,-1,1,-1,-1,1,1,1]), gl.STATIC_DRAW);
    const posLoc = gl.getAttribLocation(prog, 'aPos');
    gl.enableVertexAttribArray(posLoc);
    gl.vertexAttribPointer(posLoc, 2, gl.FLOAT, false, 0, 0);

    const uTimeLoc = gl.getUniformLocation(prog, 'uTime');
    const uResLoc  = gl.getUniformLocation(prog, 'uResolution');
    const uRandLoc = gl.getUniformLocation(prog, 'uRandom');

    const t0   = performance.now();
    const seed = Math.random();
    let animId = null;
    let firstFrame = true;

    function render() {
        const t = (performance.now() - t0) / 1000.0;
        gl.useProgram(prog);
        gl.uniform1f(uTimeLoc, t);
        gl.uniform2f(uResLoc, canvas.width, canvas.height);
        gl.uniform1f(uRandLoc, seed);
        gl.bindVertexArray(vao);
        gl.drawArrays(gl.TRIANGLE_STRIP, 0, 4);
        if (firstFrame) {
            firstFrame = false;
            console.log('[shader] first frame rendered');
            document.body.classList.add('shader-active');
        }
        animId = requestAnimationFrame(render);
    }

    document.addEventListener('visibilitychange', () => {
        if (document.hidden) { cancelAnimationFrame(animId); animId = null; }
        else if (!animId) render();
    });

    render();
})();
