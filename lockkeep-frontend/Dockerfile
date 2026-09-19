FROM node:24-alpine AS base

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME/bin:$PATH"

RUN corepack enable

WORKDIR /app

COPY package.json pnpm-lock.yaml ./

FROM base AS deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile

FROM deps AS build

COPY . .

ENV VITE_LOCKKEEP_API_URI="http://api.lockkeep.localhost/api/v1"

RUN pnpm run build

FROM deps AS runtime

WORKDIR /app

COPY --from=build /app ./

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

EXPOSE 3000

CMD ["pnpm", "run", "start"]
